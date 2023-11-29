package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

const (
	ObjectMessageType = iota
	ObjectAnnounceType
	ObjectRequestType
)

type ConstantObject struct {
	Type int
}
type MessageObject struct {
	Type    int
	Room    string
	Content string
	Author  string
}
type AnnounceObject struct {
	Type         int
	Announcement string
	Author       string
	Address      string
}
type RequestObject struct {
	Type    int
	Request string
	Author  string
	Address string
}

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

func DetermineKey(displayKey string) (realKey string) {
	for i := 0; i <= 32/len(displayKey); i++ {
		realKey += displayKey
	}
	realKey = realKey[:32]
	return realKey
}
