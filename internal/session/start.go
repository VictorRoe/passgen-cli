package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// payload is what Start sends to the agent over stdin.
type payload struct {
	Key  []byte `json:"key"`
	Salt []byte `json:"salt"`
}

// startTimeout is how long Start waits for the new agent to answer.
const startTimeout = 2 * time.Second

// RunAgent is the body of the hidden "passgen agent" command. It reads the
// key and salt from stdin, secures the socket directory, replaces any
// previous agent or stale socket, and serves until the session ends.
func RunAgent(stdin io.Reader, path string, ttl time.Duration) error {
	var p payload
	if err := json.NewDecoder(stdin).Decode(&p); err != nil {
		return fmt.Errorf("session: leer clave: %w", err)
	}
	if len(p.Key) == 0 || len(p.Salt) == 0 {
		return errors.New("session: clave o salt vacíos")
	}

	if err := secureDir(filepath.Dir(path)); err != nil {
		return err
	}

	// Newest session wins: stop a live agent, then drop any leftover socket.
	Lock(path)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("session: borrar socket anterior: %w", err)
	}

	ln, err := net.Listen("unix", path)
	if err != nil {
		return fmt.Errorf("session: abrir socket: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		ln.Close()
		return fmt.Errorf("session: permisos del socket: %w", err)
	}
	return Serve(ln, p.Key, p.Salt, ttl, time.Now)
}

// Start launches "passgen agent" detached from the terminal, hands it the
// key and salt over stdin and waits until it answers with the same salt.
func Start(key, salt []byte) error {
	path, err := SocketPath()
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("session: localizar ejecutable: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("session: obtener directorio home: %w", err)
	}
	cmd := agentCommand(exe, home)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("session: preparar agente: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("session: arrancar agente: %w", err)
	}
	encErr := json.NewEncoder(stdin).Encode(payload{Key: key, Salt: salt})
	stdin.Close()
	cmd.Process.Release()
	if encErr != nil {
		return fmt.Errorf("session: enviar clave al agente: %w", encErr)
	}

	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if _, s, err := Get(path); err == nil && bytes.Equal(s, salt) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("session: el agente no respondió")
}

// agentCommand builds the detached "passgen agent" process. It runs from the
// home directory so it never keeps the caller's folder in use (on Windows a
// folder in use cannot be deleted).
func agentCommand(exe, home string) *exec.Cmd {
	cmd := exec.Command(exe, "agent")
	cmd.Dir = home
	detach(cmd)
	return cmd
}
