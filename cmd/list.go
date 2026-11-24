package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las contraseñas almacenadas",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	v, err := openVault()
	if err != nil {
		return err
	}
	defer v.close()
	entries := v.entries

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})

	fmt.Printf("%-8s  %-32s %-20s %-32s %-16s\n", "ID", "EMAIL", "SERVICIO", "PASSWORD", "CREADO")
	fmt.Println(strings.Repeat("-", 113))
	for _, e := range entries {
		fmt.Printf("%-8s  %-32s %-20s %-32s %-16s\n",
			e.ID[:8], e.Email, e.Service, e.Password,
			e.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	return nil
}
