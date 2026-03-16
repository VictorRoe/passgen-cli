package crypto_test

import (
	"bytes"
	"testing"

	"github.com/VictorRoe/passgen-cli/internal/crypto"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("hello world — datos secretos")

	nonce, ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	got, err := crypto.Decrypt(key, nonce, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", got, plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("secret")

	nonce, ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	wrongKey := make([]byte, 32)
	wrongKey[0] = 0xFF
	_, err = crypto.Decrypt(wrongKey, nonce, ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong key: expected error, got nil")
	}
}

func TestDecrypt_AlteredCiphertext(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("secret")

	nonce, ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	ciphertext[0] ^= 0xFF
	_, err = crypto.Decrypt(key, nonce, ciphertext)
	if err == nil {
		t.Error("Decrypt() with altered ciphertext: expected error, got nil")
	}
}

func TestDeriveKey_Deterministic(t *testing.T) {
	password := []byte("mi_clave_maestra")
	salt := make([]byte, 16)

	key1 := crypto.DeriveKey(password, salt)
	key2 := crypto.DeriveKey(password, salt)

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey() not deterministic with same inputs")
	}
	if len(key1) != 32 {
		t.Errorf("DeriveKey() len = %d, want 32", len(key1))
	}
}

func TestGenerateSalt_Length(t *testing.T) {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	if len(salt) != 16 {
		t.Errorf("GenerateSalt() len = %d, want 16", len(salt))
	}
}

func TestGenerateSalt_Unique(t *testing.T) {
	s1, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() first call error = %v", err)
	}
	s2, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() second call error = %v", err)
	}
	if bytes.Equal(s1, s2) {
		t.Error("GenerateSalt() returned identical salts on two calls")
	}
}
