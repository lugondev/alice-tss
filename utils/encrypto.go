package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/getamis/sirius/log"
)

func B64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func B64Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Error("Failed to decode base64 string", "error", err)
		return nil
	}
	return data
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(text string, secret []byte, salt string) (string, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		log.Error("Encrypt NewCipher", "err", err)
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, []byte(salt)[0:16])
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return B64Encode(cipherText), nil
}

// Decrypt method is to extract back the encrypted text
func Decrypt(text string, secret []byte, salt string) (string, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	cipherText := B64Decode(text)
	cfb := cipher.NewCFBDecrypter(block, []byte(salt)[0:16])
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
