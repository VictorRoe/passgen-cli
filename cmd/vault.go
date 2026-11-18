package cmd

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
	"github.com/VictorRoe/passgen-cli/internal/storage"
)

// vault is an unlocked, decrypted vault. Call close() to zero the key.
type vault struct {
	path    string
	salt    []byte
	key     []byte
	entries []storage.Entry
}

// openVault reads the vault from its default location, prompts for the
// Master Password and unlocks it.
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

	fmt.Print("Master Password: ")
	master, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("leer master password: %w", err)
	}
	defer zero(master)

	return unlockVault(path, master)
}

// unlockVault decrypts the vault at path with master. Entries without an ID
// (written before the UUID refactor) get one, and the vault is re-saved.
func unlockVault(path string, master []byte) (*vault, error) {
	salt, nonce, ciphertext, err := storage.ReadRaw(path)
	if err != nil {
		return nil, err
	}

	key := appcrypto.DeriveKey(master, salt)
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
