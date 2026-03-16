package storage_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

func TestWriteReadRaw_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	salt := []byte("1234567890123456") // 16 bytes
	nonce := []byte("123456789012")    // 12 bytes
	ciphertext := []byte("encryptedpayload")

	if err := storage.WriteRaw(path, salt, nonce, ciphertext); err != nil {
		t.Fatalf("WriteRaw() error = %v", err)
	}

	gotSalt, gotNonce, gotCiphertext, err := storage.ReadRaw(path)
	if err != nil {
		t.Fatalf("ReadRaw() error = %v", err)
	}
	if string(gotSalt) != string(salt) {
		t.Errorf("salt mismatch: got %q want %q", gotSalt, salt)
	}
	if string(gotNonce) != string(nonce) {
		t.Errorf("nonce mismatch: got %q want %q", gotNonce, nonce)
	}
	if string(gotCiphertext) != string(ciphertext) {
		t.Errorf("ciphertext mismatch: got %q want %q", gotCiphertext, ciphertext)
	}
}

func TestReadRaw_NotExist(t *testing.T) {
	_, _, _, err := storage.ReadRaw("/nonexistent/path/vault.enc")
	if err == nil {
		t.Error("ReadRaw() on missing file: expected error, got nil")
	}
}

func TestWriteRaw_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	salt := make([]byte, 16)
	nonce := make([]byte, 12)
	ciphertext := []byte("data")

	if err := storage.WriteRaw(path, salt, nonce, ciphertext); err != nil {
		t.Fatalf("WriteRaw() error = %v", err)
	}
	// Temp file must be cleaned up after a successful write
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("WriteRaw() left .tmp file behind after successful write")
	}
	// Final file must exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("WriteRaw() final file not found: %v", err)
	}
}

func TestMarshalUnmarshalEntries_RoundTrip(t *testing.T) {
	entries := []storage.Entry{
		{
			Email:     "user@example.com",
			Service:   "github",
			Password:  "abc123XYZ!",
			CreatedAt: time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC),
		},
	}

	data, err := storage.MarshalEntries(entries)
	if err != nil {
		t.Fatalf("MarshalEntries() error = %v", err)
	}

	got, err := storage.UnmarshalEntries(data)
	if err != nil {
		t.Fatalf("UnmarshalEntries() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("UnmarshalEntries() len = %d, want 1", len(got))
	}
	if got[0].Email != entries[0].Email {
		t.Errorf("Email: got %q want %q", got[0].Email, entries[0].Email)
	}
	if got[0].Service != entries[0].Service {
		t.Errorf("Service: got %q want %q", got[0].Service, entries[0].Service)
	}
	if got[0].Password != entries[0].Password {
		t.Errorf("Password: got %q want %q", got[0].Password, entries[0].Password)
	}
	if !got[0].CreatedAt.Equal(entries[0].CreatedAt) {
		t.Errorf("CreatedAt: got %v want %v", got[0].CreatedAt, entries[0].CreatedAt)
	}
}

func TestVaultPath(t *testing.T) {
	path, err := storage.VaultPath()
	if err != nil {
		t.Fatalf("VaultPath() error = %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".passgen", "vault.enc")) {
		t.Errorf("VaultPath() got %q, want suffix .passgen/vault.enc", path)
	}
}

func TestReadRaw_CorruptedVault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := os.WriteFile(path, []byte("tooshort"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, _, _, err := storage.ReadRaw(path)
	if err == nil {
		t.Error("ReadRaw() on corrupted vault: expected error, got nil")
	}
}
