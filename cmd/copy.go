package cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

var copyCmd = &cobra.Command{
	Use:   "copy [uuid-prefix]",
	Short: "Copia una contraseña al portapapeles (se borra a los 30s)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCopy,
}

func runCopy(cmd *cobra.Command, args []string) error {
	if err := requireTerminal(os.Stdin); err != nil {
		return err
	}
	v, err := openVault()
	if err != nil {
		return err
	}
	defer v.close()

	if len(v.entries) == 0 {
		return fmt.Errorf("no hay entradas almacenadas")
	}

	var entry storage.Entry
	if len(args) == 1 {
		entry, _, err = matchEntry(v.entries, args[0])
		if err != nil {
			return err
		}
	} else {
		idx, err := runSelect("Selecciona la entrada a copiar", v.entries)
		if err != nil {
			return err
		}
		if idx < 0 {
			fmt.Println("Operación cancelada.")
			return nil
		}
		entry = v.entries[idx]
	}

	if err := clipboard.WriteAll(entry.Password); err != nil {
		return fmt.Errorf("copiar al portapapeles: %w", err)
	}

	// Clear the clipboard even if the countdown fails to start.
	countdownErr := runCountdown(entry.Service, clipboardTTL)

	cleared, err := clearIfUnchanged(clipboard.ReadAll, clipboard.WriteAll, entry.Password)
	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "⚠️  No se pudo borrar el portapapeles: %v\n", err)
	case cleared:
		fmt.Println("🧹 Portapapeles borrado.")
	default:
		fmt.Println("Portapapeles cambiado; no se borró.")
	}
	return countdownErr
}
