package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
	"github.com/VictorRoe/passgen-cli/internal/session"
	"github.com/VictorRoe/passgen-cli/internal/storage"
)

// vault is an unlocked, decrypted vault. Call close() to zero the key.
type vault struct {
	path    string
	salt    []byte
	key     []byte
	entries []storage.Entry
}

// sessionOps lets tests replace the background agent.
type sessionOps struct {
	get   func() (key, salt []byte, err error)
	start func(key, salt []byte) error
}

func realSessionOps() sessionOps {
	return sessionOps{
		get: func() ([]byte, []byte, error) {
			path, err := session.SocketPath()
			if err != nil {
				return nil, nil, err
			}
			return session.Get(path)
		},
		start: session.Start,
	}
}

// openVault unlocks the vault at its default location, using the open
// session if there is one and asking for the Master Password otherwise.
func openVault() (*vault, error) {
	path, err := storage.VaultPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("vault no encontrado, genera tu primera contraseña primero")
		}
		return nil, err
	}
	return openVaultWith(path, realSessionOps(), promptMaster)
}

// openVaultWith tries the session key first. If there is no session, it is
// for another vault (salt differs) or it does not decrypt, it asks for the
// Master Password and, on success, opens a new session. A session failure
// never blocks the command.
func openVaultWith(path string, ops sessionOps, prompt func() ([]byte, error)) (*vault, error) {
	if key, salt, err := ops.get(); err == nil {
		vaultSalt, _, _, readErr := storage.ReadRaw(path)
		if readErr == nil && bytes.Equal(salt, vaultSalt) {
			if v, err := unlockWithKey(path, key); err == nil {
				return v, nil
			}
		}
		zero(key)
	}

	master, err := prompt()
	if err != nil {
		return nil, err
	}
	defer zero(master)

	v, err := unlockVault(path, master)
	if err != nil {
		return nil, err
	}
	startSession(v, ops)
	return v, nil
}

// createVaultWith prepares a new, empty vault at path protected by master
// and opens a session for it. The caller adds entries and calls save().
func createVaultWith(path string, master []byte, ops sessionOps) (*vault, error) {
	salt, err := appcrypto.GenerateSalt()
	if err != nil {
		return nil, err
	}
	v := &vault{path: path, salt: salt, key: appcrypto.DeriveKey(master, salt)}
	startSession(v, ops)
	return v, nil
}

func startSession(v *vault, ops sessionOps) {
	if err := ops.start(v.key, v.salt); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  No se pudo abrir la sesión: %v\n", err)
		return
	}
	fmt.Println("🔓 Sesión abierta durante 15 minutos.")
}

func promptMaster() ([]byte, error) {
	fmt.Print("Master Password: ")
	master, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("leer master password: %w", err)
	}
	return master, nil
}

// unlockVault derives the key from master and unlocks the vault at path.
func unlockVault(path string, master []byte) (*vault, error) {
	salt, _, _, err := storage.ReadRaw(path)
	if err != nil {
		return nil, err
	}
	return unlockWithKey(path, appcrypto.DeriveKey(master, salt))
}

// unlockWithKey decrypts the vault at path with key. Entries without an ID
// (written before the UUID refactor) get one, and the vault is re-saved.
// On error key is zeroed.
func unlockWithKey(path string, key []byte) (*vault, error) {
	salt, nonce, ciphertext, err := storage.ReadRaw(path)
	if err != nil {
		zero(key)
		return nil, err
	}

	plaintext, err := appcrypto.Decrypt(key, nonce, ciphertext)
	if err != nil {
		zero(key)
		return nil, err
	}
	defer zero(plaintext)

	entries, err := storage.UnmarshalEntries(plaintext)
	if err != nil {
		zero(key)
		return nil, err
	}

	v := &vault{path: path, salt: salt, key: key}
	entries, changed := ensureIDs(entries)
	v.entries = entries
	if changed {
		if err := v.save(); err != nil {
			v.close()
			return nil, err
		}
	}
	return v, nil
}

// save re-encrypts the current entries and writes them to disk.
func (v *vault) save() error {
	data, err := storage.MarshalEntries(v.entries)
	if err != nil {
		return err
	}
	defer zero(data)
	nonce, ciphertext, err := appcrypto.Encrypt(v.key, data)
	if err != nil {
		return err
	}
	return storage.WriteRaw(v.path, v.salt, nonce, ciphertext)
}

func (v *vault) close() { zero(v.key) }

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
