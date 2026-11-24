package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

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
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(agentCmd)
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

	v, err := openOrCreateVault()
	if err != nil {
		return err
	}
	defer v.close()

	v.entries = append(v.entries, storage.Entry{
		ID:        uuid.NewString(),
		Email:     email,
		Service:   service,
		Password:  password,
		CreatedAt: time.Now(),
	})

	if err := v.save(); err != nil {
		return err
	}

	fmt.Println("🔐 Registro cifrado y guardado con éxito.")
	return nil
}

// openOrCreateVault unlocks the existing vault (using the session if open)
// or, on first use, asks for a new Master Password and creates the vault.
func openOrCreateVault() (*vault, error) {
	if err := storage.EnsureDir(); err != nil {
		return nil, err
	}
	path, err := storage.VaultPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		return openVaultWith(path, realSessionOps(), promptMaster)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	master, err := promptMaster()
	if err != nil {
		return nil, err
	}
	defer zero(master)
	return createVaultWith(path, master, realSessionOps())
}
