package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
	"github.com/VictorRoe/passgen-cli/internal/storage"
)

func writeTestVault(t *testing.T, master string, entries []storage.Entry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.enc")
	salt, err := appcrypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := appcrypto.DeriveKey([]byte(master), salt)
	data, err := storage.MarshalEntries(entries)
	if err != nil {
		t.Fatal(err)
	}
	nonce, ct, err := appcrypto.Encrypt(key, data)
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.WriteRaw(path, salt, nonce, ct); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUnlockVault_RightPassword(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID, Service: "github", Password: "pw1"}})
	v, err := unlockVault(path, []byte("master"))
	if err != nil {
		t.Fatalf("unlockVault() error = %v", err)
	}
	defer v.close()
	if len(v.entries) != 1 || v.entries[0].Password != "pw1" {
		t.Errorf("entries = %+v, want one entry with pw1", v.entries)
	}
}

func TestUnlockVault_WrongPassword(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID, Service: "github"}})
	before, _ := os.ReadFile(path)
	if _, err := unlockVault(path, []byte("wrong")); err == nil {
		t.Fatal("unlockVault() with wrong password must fail")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Error("vault file must not change on wrong password")
	}
}

func TestUnlockVault_AssignsMissingIDs(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{Service: "legacy"}})
	v, err := unlockVault(path, []byte("master"))
	if err != nil {
		t.Fatal(err)
	}
	id := v.entries[0].ID
	v.close()
	if len(id) != 36 {
		t.Fatalf("ID = %q, want 36-char UUID", id)
	}
	v2, err := unlockVault(path, []byte("master"))
	if err != nil {
		t.Fatal(err)
	}
	defer v2.close()
	if v2.entries[0].ID != id {
		t.Errorf("ID not persisted: got %q, want %q", v2.entries[0].ID, id)
	}
}
