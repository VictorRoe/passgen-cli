package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	saltSize  = 16
	nonceSize = 12
)

type Entry struct {
	Email     string    `json:"email"`
	Service   string    `json:"service"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func VaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("storage: obtener directorio home: %w", err)
	}
	return filepath.Join(home, ".passgen", "vault.enc"), nil
}

func EnsureDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("storage: obtener directorio home: %w", err)
	}
	dir := filepath.Join(home, ".passgen")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("storage: crear directorio: %w", err)
	}
	return nil
}

func ReadRaw(path string) (salt, nonce, ciphertext []byte, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("storage: leer vault: %w", err)
	}
	if len(data) < saltSize+nonceSize {
		return nil, nil, nil, fmt.Errorf("storage: vault corrupto (tamaño inválido)")
	}
	salt = data[:saltSize]
	nonce = data[saltSize : saltSize+nonceSize]
	ciphertext = data[saltSize+nonceSize:]
	return salt, nonce, ciphertext, nil
}

func WriteRaw(path string, salt, nonce, ciphertext []byte) error {
	tmpPath := path + ".tmp"

	content := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	content = append(content, salt...)
	content = append(content, nonce...)
	content = append(content, ciphertext...)

	if err := os.WriteFile(tmpPath, content, 0600); err != nil {
		return fmt.Errorf("storage: escribir vault temporal: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("storage: renombrar vault: %w", err)
	}
	return nil
}

func MarshalEntries(entries []Entry) ([]byte, error) {
	data, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("storage: serializar entradas: %w", err)
	}
	return data, nil
}

func UnmarshalEntries(data []byte) ([]Entry, error) {
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("storage: deserializar entradas: %w", err)
	}
	return entries, nil
}
