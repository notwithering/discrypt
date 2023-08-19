package backend

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/qeesung/image2ascii/convert"
)

func Encrypt(plainText string, key string) (string, error) {
	keyBytes := []byte(key)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	encrypter := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plainText))
	encrypter.XORKeyStream(ciphertext, []byte(plainText))

	encrypted := append(iv, ciphertext...)

	encoded := base64.URLEncoding.EncodeToString(encrypted)
	return encoded, nil
}

func Decrypt(encoded string, key string) (string, error) {
	keyBytes := []byte(key)

	encrypted, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	decrypter := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(encrypted))
	decrypter.XORKeyStream(plaintext, encrypted)

	return string(plaintext), nil
}

func CreateMessage(token, channelID, message string) error {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)

	body := []byte(fmt.Sprintf(`{"content": "%s"}`, message))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send message (status code: %d)", resp.StatusCode)
	}

	return nil
}

func GetMessages(token, channelID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve messages (status code: %d)", resp.StatusCode)
	}

	var messages []map[string]interface{}
	err = json.Unmarshal(body, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func ImageToText(filePath string, fixedWidth int) string {
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = fixedWidth
	converter := convert.NewImageConverter()
	return converter.ImageFile2ASCIIString(filePath, &convertOptions)
}

func DownloadImage(url string) (string, error) {
	tempDir, err := ioutil.TempDir("", "image-download")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download the image: %v", err)
	}
	defer response.Body.Close()

	file, err := ioutil.TempFile(tempDir, "image-*.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save the image: %v", err)
	}

	filePath := file.Name()

	return filePath, nil
}
