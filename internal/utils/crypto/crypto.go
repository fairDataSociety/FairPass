package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type Crypto interface {
	EncryptContent(string, string) (string, error)
	DecryptContent(string, string) (string, error)
}

type Encryptor struct{}

func (*Encryptor) DecryptContent(passphrase, encryptedContent string) (string, error) {
	password := passphrase
	if password == "" {
		return "", fmt.Errorf("passphrase cannot be blank")
	}

	if encryptedContent == "" {
		return "", fmt.Errorf("invalid encrypted content")
	}
	aesKey := sha256.Sum256([]byte(password))

	//decrypt the message
	data, err := decrypt(aesKey[:], encryptedContent)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (*Encryptor) EncryptContent(passphrase, data string) (string, error) {
	password := passphrase
	if password == "" {
		return "", fmt.Errorf("passphrase cannot be blank")
	}
	aesKey := sha256.Sum256([]byte(password))
	encryptedMessage, err := encrypt(aesKey[:], data)
	if err != nil {
		return "", fmt.Errorf("create user account: %w", err)
	}
	return encryptedMessage, nil
}

func encrypt(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

func decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = fmt.Errorf("ciphertext block size is too short")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}
