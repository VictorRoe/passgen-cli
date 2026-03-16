package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"runtime"

	"golang.org/x/crypto/argon2"
)

const (
	saltSize  = 16
	nonceSize = 12
	keySize   = 32

	argonTime   uint32 = 3
	argonMemory uint32 = 64 * 1024 // 64 MB in KB
)

func DeriveKey(password, salt []byte) []byte {
	threads := uint8(runtime.NumCPU())
	if threads > 4 {
		threads = 4
	}
	return argon2.IDKey(password, salt, argonTime, argonMemory, threads, keySize)
}

// GenerateSalt returns 16 cryptographically random bytes.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("crypto: generar salt: %w", err)
	}
	return salt, nil
}

// Encrypt encrypts plaintext with AES-256-GCM using the given 32-byte key.
// Returns a freshly generated nonce and the ciphertext (with GCM tag appended).
func Encrypt(key, plaintext []byte) (nonce []byte, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("crypto: crear cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("crypto: crear GCM: %w", err)
	}
	nonce = make([]byte, nonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("crypto: generar nonce: %w", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts ciphertext with AES-256-GCM.
// Returns a descriptive error if authentication fails (wrong key or tampered data).
func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: crear cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: crear GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("master password incorrecta o archivo corrupto")
	}
	return plaintext, nil
}
