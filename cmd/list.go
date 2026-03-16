package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
	"github.com/VictorRoe/passgen-cli/internal/storage"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las contraseñas almacenadas",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
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

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})

	fmt.Printf("%-32s %-20s %-32s %-16s\n", "EMAIL", "SERVICIO", "PASSWORD", "CREADO")
	fmt.Println(strings.Repeat("-", 104))
	for _, e := range entries {
		fmt.Printf("%-32s %-20s %-32s %-16s\n",
			e.Email, e.Service, e.Password,
			e.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	return nil
}
