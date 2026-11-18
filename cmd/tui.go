package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

// confirmGuard is how long Enter is ignored after the confirmation appears,
// so a buffered keypress (double Enter at the password or in the selector)
// cannot delete an entry without a deliberate press.
const confirmGuard = 300 * time.Millisecond

// confirmModel asks for a single-key confirmation: Enter confirms,
// Esc or Ctrl+C cancels, every other key is ignored.
type confirmModel struct {
	entry     storage.Entry
	confirmed bool
	shownAt   time.Time
	now       func() time.Time
}

func newConfirmModel(e storage.Entry) confirmModel {
	return confirmModel{entry: e, shownAt: time.Now(), now: time.Now}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.Type {
	case tea.KeyEnter:
		if m.now().Sub(m.shownAt) < confirmGuard {
			return m, nil
		}
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
	title   string
	entries []storage.Entry
	cursor  int
	chosen  int
}

func newSelectModel(title string, entries []storage.Entry) selectModel {
	return selectModel{title: title, entries: entries, chosen: -1}
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
	fmt.Fprintf(&b, "%s (↑/↓, Enter, Esc/Ctrl+C para salir):\n", m.title)
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
func runSelect(title string, entries []storage.Entry) (int, error) {
	final, err := tea.NewProgram(newSelectModel(title, entries)).Run()
	if err != nil {
		return -1, fmt.Errorf("selector: %w", err)
	}
	return final.(selectModel).chosen, nil
}

// requireTerminal fails when f is not an interactive terminal. Bubble Tea
// would otherwise open the console directly and still read keys from it.
func requireTerminal(f *os.File) error {
	if !term.IsTerminal(int(f.Fd())) {
		return fmt.Errorf("delete requiere una terminal interactiva")
	}
	return nil
}

// tickMsg is sent once per second while the countdown runs.
type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

// countdownModel shows how long until the clipboard is cleared. It quits at
// zero or when Enter, Esc or Ctrl+C is pressed. It never sees the password.
type countdownModel struct {
	service   string
	remaining int
}

func newCountdownModel(service string, seconds int) countdownModel {
	return countdownModel{service: service, remaining: seconds}
}

func (m countdownModel) Init() tea.Cmd { return tick() }

func (m countdownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.remaining--
		if m.remaining <= 0 {
			return m, tea.Quit
		}
		return m, tick()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m countdownModel) View() string {
	return fmt.Sprintf("✅ Password de '%s' copiado. Se borrará en %ds · [Enter/Esc/Ctrl+C] borrar ahora\n",
		m.service, m.remaining)
}

// runCountdown blocks until the countdown ends or the user ends it early.
func runCountdown(service string, seconds int) error {
	if _, err := tea.NewProgram(newCountdownModel(service, seconds)).Run(); err != nil {
		return fmt.Errorf("cuenta atrás: %w", err)
	}
	return nil
}
