package cmd

import (
	"fmt"
	"strings"

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

// selectModel lets the user pick one entry with ↑/↓ (or k/j).
// Enter chooses the highlighted entry; Esc or Ctrl+C cancels (chosen == -1).
type selectModel struct {
	entries []storage.Entry
	cursor  int
	chosen  int
}

func newSelectModel(entries []storage.Entry) selectModel {
	return selectModel{entries: entries, chosen: -1}
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter":
		m.chosen = m.cursor
		return m, tea.Quit
	case "esc", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m selectModel) View() string {
	var b strings.Builder
	b.WriteString("Selecciona la entrada a eliminar (↑/↓, Enter, Esc/Ctrl+C para salir):\n")
	for i, e := range m.entries {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		fmt.Fprintf(&b, "%s%-8s  %-20s  (%s)\n", prefix, e.ID[:8], e.Service, e.Email)
	}
	return b.String()
}

// runSelect shows the selector and returns the chosen index, or -1 if cancelled.
func runSelect(entries []storage.Entry) (int, error) {
	final, err := tea.NewProgram(newSelectModel(entries)).Run()
	if err != nil {
		return -1, fmt.Errorf("selector: %w", err)
	}
	return final.(selectModel).chosen, nil
}
