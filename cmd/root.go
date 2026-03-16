package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	appcrypto "github.com/VictorRoe/passgen-cli/internal/crypto"
	"github.com/VictorRoe/passgen-cli/internal/generator"
	"github.com/VictorRoe/passgen-cli/internal/storage"
)

var (
	flagLength  int
	flagSymbols bool
	flagNumbers bool
)

var rootCmd = &cobra.Command{
	Use:   "passgen <email> <service>",
	Short: "Genera y guarda una contraseña cifrada",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerate,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&flagLength, "chars", "C", 16, "Longitud del password (mínimo 8)")
	rootCmd.Flags().BoolVarP(&flagSymbols, "symbols", "S", false, "Incluir símbolos (@#$%&*()!)")
	rootCmd.Flags().BoolVarP(&flagNumbers, "numbers", "N", false, "Incluir números (0-9)")
	rootCmd.AddCommand(listCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	email, service := args[0], args[1]

	if flagLength < 8 {
		return fmt.Errorf("la longitud mínima es 8 caracteres")
	}

	opts := generator.Options{
		Length:  flagLength,
		Symbols: flagSymbols,
		Numbers: flagNumbers,
	}
	password, err := generator.Generate(opts)
	if err != nil {
		return err
	}

	if err := clipboard.WriteAll(password); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  No se pudo copiar al portapapeles: %v\n", err)
	} else {
		fmt.Println("✅ Password copiado al portapapeles.")
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

	if err := storage.EnsureDir(); err != nil {
		return err
	}

	vaultPath, err := storage.VaultPath()
	if err != nil {
		return err
	}

	var salt []byte
	var entries []storage.Entry
	isNewVault := false

	rawSalt, nonce, ciphertext, readErr := storage.ReadRaw(vaultPath)
	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return readErr
		}
		isNewVault = true
		salt, err = appcrypto.GenerateSalt()
		if err != nil {
			return err
		}
	} else {
		salt = rawSalt
	}

	key := appcrypto.DeriveKey(masterBytes, salt)
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()

	if !isNewVault {
		plaintext, err := appcrypto.Decrypt(key, nonce, ciphertext)
		if err != nil {
			return err
		}
		defer func() {
			for i := range plaintext {
				plaintext[i] = 0
			}
		}()
		entries, err = storage.UnmarshalEntries(plaintext)
		if err != nil {
			return err
		}
	}

	entries = append(entries, storage.Entry{
		Email:     email,
		Service:   service,
		Password:  password,
		CreatedAt: time.Now(),
	})

	jsonData, err := storage.MarshalEntries(entries)
	if err != nil {
		return err
	}
	newNonce, newCiphertext, err := appcrypto.Encrypt(key, jsonData)
	if err != nil {
		return err
	}

	if err := storage.WriteRaw(vaultPath, salt, newNonce, newCiphertext); err != nil {
		return err
	}

	fmt.Println("🔐 Registro cifrado y guardado con éxito.")
	return nil
}
