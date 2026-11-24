package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/VictorRoe/passgen-cli/internal/session"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Cierra la sesión: la Master Password se volverá a pedir",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := session.SocketPath()
		if err != nil {
			return err
		}
		closed, err := session.Lock(path)
		if err != nil {
			return err
		}
		if closed {
			fmt.Println("🔒 Sesión cerrada.")
		} else {
			fmt.Println("No hay sesión abierta.")
		}
		return nil
	},
}

// agentCmd is the background process that holds the session key. It is
// started by passgen itself and receives the key over stdin.
var agentCmd = &cobra.Command{
	Use:    "agent",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := session.SocketPath()
		if err != nil {
			return err
		}
		return session.RunAgent(os.Stdin, path, session.TTL)
	},
}
