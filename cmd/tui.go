package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

// confirmModel asks for a single-key confirmation: Enter confirms,
// Esc or Ctrl+C cancels, every other key is ignored.
type confirmModel struct {
	entry     storage.Entry
	confirmed bool
}

func newConfirmModel(e storage.Entry) confirmModel {
	return confirmModel{entry: e}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.Type {
	case tea.KeyEnter:
		m.confirmed = true
		return m, tea.Quit
	case tea.KeyEsc, tea.KeyCtrlC:
		return m, tea.Quit
	}
	return m, nil
}

func (m confirmModel) View() string {
	return fmt.Sprintf("¿Eliminar '%s' (%s)? [Enter] confirmar · [Esc/Ctrl+C] cancelar\n",
		m.entry.Service, m.entry.Email)
}

// runConfirm shows the confirmation prompt and reports whether the user confirmed.
func runConfirm(e storage.Entry) (bool, error) {
	final, err := tea.NewProgram(newConfirmModel(e)).Run()
	if err != nil {
		return false, fmt.Errorf("confirmación: %w", err)
	}
	return final.(confirmModel).confirmed, nil
}
