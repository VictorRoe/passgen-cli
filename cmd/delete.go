package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/VictorRoe/passgen-cli/internal/storage"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
)

// ensureIDs is a pure function that assigns a UUID v4 to any Entry whose ID
// is empty (legacy entries written before the UUID refactor). It returns the
// updated slice and a boolean indicating whether any IDs were newly assigned.
// Invariant: after this call every entry has a 36-character UUID string as ID.
func ensureIDs(entries []storage.Entry) ([]storage.Entry, bool) {
	changed := false
	for i := range entries {
		if entries[i].ID == "" {
			entries[i].ID = uuid.NewString()
			changed = true
		}
	}
	return entries, changed
}

// matchEntry finds the unique entry whose ID starts with prefix.
// Returns the entry, its index in the slice, and an error if the prefix
// is too short (< 8 chars), matches nothing, or is ambiguous.
func matchEntry(entries []storage.Entry, prefix string) (storage.Entry, int, error) {
	if len(prefix) < 8 {
		return storage.Entry{}, -1, fmt.Errorf("el prefijo debe tener al menos 8 caracteres")
	}
	var matches []int
	for i, e := range entries {
		if strings.HasPrefix(e.ID, prefix) {
			matches = append(matches, i)
		}
	}
	switch len(matches) {
	case 0:
		return storage.Entry{}, -1, fmt.Errorf("no se encontró ninguna entrada con ID %q", prefix)
	case 1:
		return entries[matches[0]], matches[0], nil
	default:
		return storage.Entry{}, -1, fmt.Errorf("el prefijo %q coincide con más de una entrada, usa más caracteres", prefix)
	}
}

var deleteCmd = &cobra.Command{
	Use:   "delete [uuid-prefix]",
	Short: "Elimina una contraseña almacenada por su ID",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	vaultPath, err := storage.VaultPath()
	if err != nil {
		return err
	}

	salt, nonce, ciphertext, err := storage.ReadRaw(vaultPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("vault no encontrado, genera tu primera contraseña primero")
		}
		return err
	}

	if err := requireTerminal(os.Stdin); err != nil {
		return err
	}

	fmt.Print("Master Password: ")
	masterBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("leer master password: %w", err)
	}
	defer func() {
		for i := range masterBytes {
			masterBytes[i] = 0
		}
	}()

	key := appcrypto.DeriveKey(masterBytes, salt)
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()

	plaintext, err := appcrypto.Decrypt(key, nonce, ciphertext)
	if err != nil {
		return err
	}
	defer func() {
		for i := range plaintext {
			plaintext[i] = 0
		}
	}()

	entries, err := storage.UnmarshalEntries(plaintext)
	if err != nil {
		return err
	}

	entries, changed := ensureIDs(entries)
	if changed {
		jsonData, marshalErr := storage.MarshalEntries(entries)
		if marshalErr != nil {
			return marshalErr
		}
		defer func() {
			for i := range jsonData {
				jsonData[i] = 0
			}
		}()
		newNonce, newCiphertext, encErr := appcrypto.Encrypt(key, jsonData)
		if encErr != nil {
			return encErr
		}
		if writeErr := storage.WriteRaw(vaultPath, salt, newNonce, newCiphertext); writeErr != nil {
			return writeErr
		}
	}

	if len(entries) == 0 {
		return fmt.Errorf("no hay entradas almacenadas")
	}

	if len(args) == 1 {
		return deleteByPrefix(entries, args[0], key, salt, vaultPath)
	}
	return deleteInteractive(entries, key, salt, vaultPath)
}

func deleteByPrefix(entries []storage.Entry, prefix string, key, salt []byte, vaultPath string) error {
	entry, idx, err := matchEntry(entries, prefix)
	if err != nil {
		return err
	}

	confirmed, err := runConfirm(entry)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operación cancelada.")
		return nil
	}

	return removeAndSave(entries, idx, key, salt, vaultPath)
}

func removeAndSave(entries []storage.Entry, idx int, key, salt []byte, vaultPath string) error {
	entries = append(entries[:idx], entries[idx+1:]...)

	jsonData, err := storage.MarshalEntries(entries)
	if err != nil {
		return err
	}
	defer func() {
		for i := range jsonData {
			jsonData[i] = 0
		}
	}()

	newNonce, newCiphertext, err := appcrypto.Encrypt(key, jsonData)
	if err != nil {
		return err
	}
	if err := storage.WriteRaw(vaultPath, salt, newNonce, newCiphertext); err != nil {
		return err
	}

	fmt.Println("Entrada eliminada.")
	return nil
}

func deleteInteractive(entries []storage.Entry, key, salt []byte, vaultPath string) error {
	idx, err := runSelect(entries)
	if err != nil {
		return err
	}
	if idx < 0 {
		fmt.Println("Operación cancelada.")
		return nil
	}

	confirmed, err := runConfirm(entries[idx])
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operación cancelada.")
		return nil
	}

	return removeAndSave(entries, idx, key, salt, vaultPath)
}
