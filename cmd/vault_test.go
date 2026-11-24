package cmd

import (
	"bytes"
	"errors"
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

// fakeSession records how openVaultWith uses the session.
type fakeSession struct {
	key, salt  []byte
	getErr     error
	startErr   error
	started    bool
	startedKey []byte
	startSalt  []byte
}

func (f *fakeSession) ops() sessionOps {
	return sessionOps{
		get: func() ([]byte, []byte, error) {
			if f.getErr != nil {
				return nil, nil, f.getErr
			}
			return append([]byte(nil), f.key...), append([]byte(nil), f.salt...), nil
		},
		start: func(key, salt []byte) error {
			f.started = true
			f.startedKey = append([]byte(nil), key...)
			f.startSalt = append([]byte(nil), salt...)
			return f.startErr
		},
	}
}

func vaultKeyAndSalt(t *testing.T, path, master string) ([]byte, []byte) {
	t.Helper()
	salt, _, _, err := storage.ReadRaw(path)
	if err != nil {
		t.Fatal(err)
	}
	return appcrypto.DeriveKey([]byte(master), salt), salt
}

func promptReturning(t *testing.T, pw string, called *bool) func() ([]byte, error) {
	return func() ([]byte, error) {
		*called = true
		return []byte(pw), nil
	}
}

func TestUnlockFromSession_UsesSessionKey(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID, Password: "pw1"}})
	key, salt := vaultKeyAndSalt(t, path, "master")
	fs := &fakeSession{key: key, salt: salt}
	prompted := false
	v, err := openVaultWith(path, fs.ops(), promptReturning(t, "master", &prompted))
	if err != nil {
		t.Fatalf("openVaultWith() error = %v", err)
	}
	defer v.close()
	if prompted {
		t.Error("Master Password must not be asked while the session is valid")
	}
	if fs.started {
		t.Error("a valid session must not be restarted")
	}
	if v.entries[0].Password != "pw1" {
		t.Errorf("entries = %+v", v.entries)
	}
}

func TestUnlockFromSession_SaltMismatch(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID}})
	key, _ := vaultKeyAndSalt(t, path, "master")
	fs := &fakeSession{key: key, salt: []byte("another-vault-salt")}
	prompted := false
	v, err := openVaultWith(path, fs.ops(), promptReturning(t, "master", &prompted))
	if err != nil {
		t.Fatal(err)
	}
	defer v.close()
	if !prompted {
		t.Error("a session for another vault must be ignored and the password asked")
	}
}

func TestUnlockFromSession_BadKey(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID}})
	_, salt := vaultKeyAndSalt(t, path, "master")
	fs := &fakeSession{key: make([]byte, 32), salt: salt}
	prompted := false
	v, err := openVaultWith(path, fs.ops(), promptReturning(t, "master", &prompted))
	if err != nil {
		t.Fatal(err)
	}
	defer v.close()
	if !prompted {
		t.Error("a session key that does not decrypt must be ignored")
	}
}

func TestOpenVaultWith_StartsSessionAfterPrompt(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID}})
	key, salt := vaultKeyAndSalt(t, path, "master")
	fs := &fakeSession{getErr: errors.New("no session")}
	prompted := false
	v, err := openVaultWith(path, fs.ops(), promptReturning(t, "master", &prompted))
	if err != nil {
		t.Fatal(err)
	}
	defer v.close()
	if !prompted || !fs.started {
		t.Fatalf("prompted=%v started=%v, want both true", prompted, fs.started)
	}
	if !bytes.Equal(fs.startedKey, key) || !bytes.Equal(fs.startSalt, salt) {
		t.Error("session must be started with the vault's key and salt")
	}
}

func TestOpenVaultWith_WrongPasswordNoSession(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID}})
	fs := &fakeSession{getErr: errors.New("no session")}
	prompted := false
	if _, err := openVaultWith(path, fs.ops(), promptReturning(t, "wrong", &prompted)); err == nil {
		t.Fatal("wrong password must fail")
	}
	if fs.started {
		t.Error("a wrong password must never open a session")
	}
}

func TestOpenVaultWith_StartFailureStillUnlocks(t *testing.T) {
	path := writeTestVault(t, "master", []storage.Entry{{ID: testEntry.ID, Password: "pw1"}})
	fs := &fakeSession{getErr: errors.New("no session"), startErr: errors.New("agent failed")}
	prompted := false
	v, err := openVaultWith(path, fs.ops(), promptReturning(t, "master", &prompted))
	if err != nil {
		t.Fatalf("a session failure must not block the command: %v", err)
	}
	defer v.close()
	if v.entries[0].Password != "pw1" {
		t.Errorf("entries = %+v", v.entries)
	}
}

func TestCreateVaultWith_OpensSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	fs := &fakeSession{}
	v, err := createVaultWith(path, []byte("master"), fs.ops())
	if err != nil {
		t.Fatal(err)
	}
	defer v.close()
	if !fs.started || len(fs.startSalt) == 0 {
		t.Error("creating a vault must open a session")
	}
	if err := v.save(); err != nil {
		t.Fatal(err)
	}
	v2, err := unlockVault(path, []byte("master"))
	if err != nil {
		t.Fatalf("new vault must unlock with its master password: %v", err)
	}
	v2.close()
}
