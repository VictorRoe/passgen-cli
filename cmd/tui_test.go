package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

var testEntry = storage.Entry{ID: "abcdefgh-0000-0000-0000-000000000001", Service: "github", Email: "a@b.com"}

func key(t tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: t} }
func runeKey(r rune) tea.KeyMsg   { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func TestConfirm_EnterConfirms(t *testing.T) {
	m, cmd := newConfirmModel(testEntry).Update(key(tea.KeyEnter))
	if !m.(confirmModel).confirmed {
		t.Error("Enter must confirm")
	}
	if cmd == nil {
		t.Error("Enter must quit the program")
	}
}

func TestConfirm_EscCancels(t *testing.T) {
	m, cmd := newConfirmModel(testEntry).Update(key(tea.KeyEsc))
	if m.(confirmModel).confirmed || cmd == nil {
		t.Error("Esc must cancel and quit")
	}
}

func TestConfirm_CtrlCCancels(t *testing.T) {
	m, cmd := newConfirmModel(testEntry).Update(key(tea.KeyCtrlC))
	if m.(confirmModel).confirmed || cmd == nil {
		t.Error("Ctrl+C must cancel and quit")
	}
}

func TestConfirm_OtherKeysIgnored(t *testing.T) {
	for _, r := range []rune{'s', 'y', ' '} {
		m, cmd := newConfirmModel(testEntry).Update(runeKey(r))
		if m.(confirmModel).confirmed || cmd != nil {
			t.Errorf("key %q must be ignored", r)
		}
	}
}

func TestConfirm_ViewShowsEntry(t *testing.T) {
	v := newConfirmModel(testEntry).View()
	if !strings.Contains(v, "github") || !strings.Contains(v, "a@b.com") {
		t.Errorf("View() = %q, want service and email", v)
	}
}
