package midend

import (
	"main/backend"
)

func SendMessage(token, channelID, message, encryptionKey string) {
	if len(encryptionKey) == 4 {
		encryptionKey += encryptionKey // 8
		encryptionKey += encryptionKey // 16
		encryptionKey += encryptionKey // 32
	} else if len(encryptionKey) == 8 {
		encryptionKey += encryptionKey // 16
		encryptionKey += encryptionKey // 32
	}
	encrypted, err := backend.Encrypt(message, encryptionKey)
	if err != nil {
		panic(err)
	}
	backend.CreateMessage(token, channelID, encrypted)
}

func ImageURLToText(url string, fixedWidth int) (string, error) {
	path, err := backend.DownloadImage(url)
	if err != nil {
		return "", err
	}
	return backend.ImageToText(path, fixedWidth), nil
}
