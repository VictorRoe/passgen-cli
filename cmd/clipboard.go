package cmd

import "fmt"

// clipboardTTL is how many seconds a copied password stays in the clipboard.
const clipboardTTL = 30

// clearIfUnchanged empties the clipboard only if it still holds password,
// so anything the user copied in the meantime is left alone. It reports
// whether the clipboard was cleared.
func clearIfUnchanged(read func() (string, error), write func(string) error, password string) (bool, error) {
	current, err := read()
	if err != nil {
		return false, fmt.Errorf("leer portapapeles: %w", err)
	}
	if current != password {
		return false, nil
	}
	if err := write(""); err != nil {
		return false, fmt.Errorf("borrar portapapeles: %w", err)
	}
	return true, nil
}
