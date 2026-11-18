package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/VictorRoe/passgen-cli/internal/storage"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
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
	if len(args) == 1 {
		return deleteByPrefix(v, args[0])
	}
	return deleteInteractive(v)
}

func deleteByPrefix(v *vault, prefix string) error {
	entry, idx, err := matchEntry(v.entries, prefix)
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

	return removeAndSave(v, idx)
}

func removeAndSave(v *vault, idx int) error {
	v.entries = append(v.entries[:idx], v.entries[idx+1:]...)
	if err := v.save(); err != nil {
		return err
	}
	fmt.Println("Entrada eliminada.")
	return nil
}

func deleteInteractive(v *vault) error {
	idx, err := runSelect("Selecciona la entrada a eliminar", v.entries)
	if err != nil {
		return err
	}
	if idx < 0 {
		fmt.Println("Operación cancelada.")
		return nil
	}

	confirmed, err := runConfirm(v.entries[idx])
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Operación cancelada.")
		return nil
	}

	return removeAndSave(v, idx)
}
