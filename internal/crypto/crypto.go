package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Cipher wraps AES-256-GCM for envelope encryption of sensitive fields.
type Cipher struct {
	gcm cipher.AEAD
}

// NewCipher creates a Cipher from a base64-encoded 32-byte key.
// Generate a key with: openssl rand -base64 32
func NewCipher(keyB64 string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes (base64 encoded)")
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

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext (nonce + ciphertext).
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext.
func (c *Cipher) Decrypt(cipherB64 string) (string, error) {
	if cipherB64 == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := c.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// MaskPhone masks a phone number: 13812345678 → 138****5678.
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskCreditCode masks a credit code: shows first 4 and last 4 chars.
func MaskCreditCode(code string) string {
	if len(code) < 8 {
		return strings.Repeat("*", len(code))
	}
	return code[:4] + strings.Repeat("*", len(code)-8) + code[len(code)-4:]
}

// MaskIDCard masks an ID card number: shows first 3 and last 4 chars.
func MaskIDCard(id string) string {
	if len(id) < 7 {
		return strings.Repeat("*", len(id))
	}
	return id[:3] + strings.Repeat("*", len(id)-7) + id[len(id)-4:]
}
