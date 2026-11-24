// Package session keeps the derived vault key in a background agent's
// memory for a fixed time, so the Master Password is asked only once.
package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TTL is how long a session lasts from the moment it is opened. It is never renewed.
const TTL = 15 * time.Minute

// ErrNoSession means there is no usable session (no agent, stale socket or expired).
var ErrNoSession = errors.New("no hay sesión abierta")

// ioTimeout bounds every client/agent exchange.
const ioTimeout = time.Second

type request struct {
	Op string `json:"op"`
}

type response struct {
	Key       []byte    `json:"key,omitempty"`
	Salt      []byte    `json:"salt,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	OK        bool      `json:"ok,omitempty"`
	Err       string    `json:"err,omitempty"`
}

// SocketPath returns ~/.passgen/agent.sock.
func SocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("session: obtener directorio home: %w", err)
	}
	return filepath.Join(home, ".passgen", "agent.sock"), nil
}
