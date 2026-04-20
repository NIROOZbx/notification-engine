package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/NIROOZbx/notification-engine/pkg/serializer"
)

func EncryptMap(data map[string]string, secretKey string) (string, error) {
	bytes, err := serializer.Marshal(data)
	if err != nil {
		return "", err
	}
	return Encrypt(bytes, secretKey)
}

func DecryptToMap(encrypted string, secretKey string) (map[string]string, error) {
	bytes, err := Decrypt(encrypted, secretKey)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	if err := serializer.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func Encrypt(plaintext []byte, secretKey string) (string, error) {

	key := []byte(secretKey)
	if len(key) != 32 {
		fmt.Println(len(key))
		return "", errors.New("encryption key must be exactly 32 bytes long")
	}


	block,err:=aes.NewCipher(key)

	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(cipherTextStr string, secretKey string) ([]byte, error) {
	key := []byte(secretKey)
	if len(key) != 32 {
		return nil, errors.New("encryption key must be exactly 32 bytes long")
	}

	data, err := base64.StdEncoding.DecodeString(cipherTextStr)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
