//go:build !windows

package session

import (
	"fmt"
	"os"
)

// secureDir makes dir accessible only to the current user.
func secureDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("session: crear directorio: %w", err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("session: permisos del directorio: %w", err)
	}
	return nil
}
