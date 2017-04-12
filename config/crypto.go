package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/sha3"
)

/**
 * https://gist.github.com/josephspurrier/12cc5ed76d2228a41ceb
 */

func decrypt(cipherstring string, keystring string) string {
	ciphertext, _ := base64.StdEncoding.DecodeString(cipherstring)
	key := sha3.Sum256([]byte(keystring))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("Text is too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}

func encrypt(plainstring, keystring string) string {
	plaintext := []byte(plainstring)
	key := sha3.Sum256([]byte(keystring))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}
