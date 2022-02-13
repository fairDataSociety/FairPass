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

type Encryptor struct {
	encryptedMnemonic string
}

func New(encryptedMnemonic string) *Encryptor {
	return &Encryptor{encryptedMnemonic}
}

func (a *Encryptor) GetEncryptedMnemonic() string {
	return a.encryptedMnemonic
}

func (a *Encryptor) EncryptContent(password, content string) (string, error) {
	mnemonic, err := a.decryptMnemonic(password)
	if err != nil {
		return "", fmt.Errorf("load mnemonic: %w", err)
	}
	aesKey := sha256.Sum256([]byte(mnemonic))

	// encrypt the mnemonic
	encryptedContent, err := encrypt(aesKey[:], content)
	if err != nil {
		return "", fmt.Errorf("load mnemonic: %w", err)
	}

	return encryptedContent, nil
}

func (a *Encryptor) DecryptContent(password, content string) (string, error) {
	mnemonic, err := a.decryptMnemonic(password)
	if err != nil {
		return "", fmt.Errorf("load mnemonic: %w", err)
	}
	aesKey := sha256.Sum256([]byte(mnemonic))

	//decrypt the message
	decryptedContent, err := decrypt(aesKey[:], content)
	if err != nil {
		return "", err
	}

	return decryptedContent, nil
}

func (a *Encryptor) decryptMnemonic(password string) (string, error) {
	aesKey := sha256.Sum256([]byte(password))

	//decrypt the message
	mnemonic, err := decrypt(aesKey[:], a.encryptedMnemonic)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
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

	//returns to base64 encoded string
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
