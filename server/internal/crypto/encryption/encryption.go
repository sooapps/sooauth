package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const prefix = "v1:"

var (
	ErrInvalidKey        = errors.New("encryption key must be 32 bytes")
	ErrInvalidCiphertext = errors.New("invalid encrypted value")
)

type Cipher struct{ gcm cipher.AEAD }

func New(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &Cipher{gcm: gcm}, nil
}

func (c *Cipher) Encrypt(plain string) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := c.gcm.Seal(nil, nonce, []byte(plain), []byte(prefix))
	return prefix + base64.RawStdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func (c *Cipher) Decrypt(value string) (string, error) {
	if !IsEncrypted(value) {
		return "", ErrInvalidCiphertext
	}
	data, err := base64.RawStdEncoding.DecodeString(value[len(prefix):])
	if err != nil || len(data) < c.gcm.NonceSize()+c.gcm.Overhead() {
		return "", ErrInvalidCiphertext
	}
	plain, err := c.gcm.Open(nil, data[:c.gcm.NonceSize()], data[c.gcm.NonceSize():], []byte(prefix))
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plain), nil
}

func IsEncrypted(value string) bool {
	return len(value) > len(prefix) && value[:len(prefix)] == prefix
}
