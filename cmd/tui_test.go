package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

var testEntry = storage.Entry{ID: "abcdefgh-0000-0000-0000-000000000001", Service: "github", Email: "a@b.com"}

func key(t tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: t} }
func runeKey(r rune) tea.KeyMsg    { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func TestConfirm_EnterConfirms(t *testing.T) {
	c := newConfirmModel(testEntry)
	c.shownAt = time.Now().Add(-confirmGuard)
	m, cmd := c.Update(key(tea.KeyEnter))
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

func threeEntries() []storage.Entry {
	return []storage.Entry{
		{ID: "11111111-0000-0000-0000-000000000001", Service: "github", Email: "a@b.com"},
		{ID: "22222222-0000-0000-0000-000000000002", Service: "gitlab", Email: "c@d.com"},
		{ID: "33333333-0000-0000-0000-000000000003", Service: "aws", Email: "e@f.com"},
	}
}

func sendAll(m tea.Model, msgs ...tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		m, cmd = m.Update(msg)
	}
	return m, cmd
}

func TestSelect_Navigation(t *testing.T) {
	m, _ := sendAll(newSelectModel(threeEntries()), key(tea.KeyDown), key(tea.KeyDown), key(tea.KeyUp))
	if got := m.(selectModel).cursor; got != 1 {
		t.Errorf("arrows: cursor = %d, want 1", got)
	}
	m, _ = sendAll(newSelectModel(threeEntries()), runeKey('j'), runeKey('j'), runeKey('k'))
	if got := m.(selectModel).cursor; got != 1 {
		t.Errorf("j/k: cursor = %d, want 1", got)
	}
}

func TestSelect_Bounds(t *testing.T) {
	m, _ := sendAll(newSelectModel(threeEntries()), key(tea.KeyUp))
	if got := m.(selectModel).cursor; got != 0 {
		t.Errorf("up at top: cursor = %d, want 0", got)
	}
	m, _ = sendAll(newSelectModel(threeEntries()), key(tea.KeyDown), key(tea.KeyDown), key(tea.KeyDown), key(tea.KeyDown))
	if got := m.(selectModel).cursor; got != 2 {
		t.Errorf("down at bottom: cursor = %d, want 2", got)
	}
}

func TestSelect_EnterChooses(t *testing.T) {
	m, cmd := sendAll(newSelectModel(threeEntries()), key(tea.KeyDown), key(tea.KeyEnter))
	if got := m.(selectModel).chosen; got != 1 {
		t.Errorf("chosen = %d, want 1", got)
	}
	if cmd == nil {
		t.Error("Enter must quit the program")
	}
}

func TestSelect_EscCancels(t *testing.T) {
	m, cmd := sendAll(newSelectModel(threeEntries()), key(tea.KeyDown), key(tea.KeyEsc))
	if m.(selectModel).chosen != -1 || cmd == nil {
		t.Error("Esc must cancel (chosen=-1) and quit")
	}
}

func TestSelect_CtrlCCancels(t *testing.T) {
	m, cmd := sendAll(newSelectModel(threeEntries()), key(tea.KeyCtrlC))
	if m.(selectModel).chosen != -1 || cmd == nil {
		t.Error("Ctrl+C must cancel (chosen=-1) and quit")
	}
}

func TestSelect_ViewMarksCursor(t *testing.T) {
	m, _ := sendAll(newSelectModel(threeEntries()), key(tea.KeyDown))
	var active []string
	for _, line := range strings.Split(m.View(), "\n") {
		if strings.HasPrefix(line, "> ") {
			active = append(active, line)
		}
	}
	if len(active) != 1 {
		t.Fatalf("want exactly one active line, got %d in %q", len(active), m.View())
	}
	for _, want := range []string{"22222222", "gitlab", "c@d.com"} {
		if !strings.Contains(active[0], want) {
			t.Errorf("active line %q missing %q", active[0], want)
		}
	}
}

func TestConfirm_EnterIgnoredRightAfterShown(t *testing.T) {
	start := time.Now()
	m := newConfirmModel(testEntry)
	m.shownAt = start
	m.now = func() time.Time { return start.Add(confirmGuard / 2) }
	got, cmd := m.Update(key(tea.KeyEnter))
	if got.(confirmModel).confirmed || cmd != nil {
		t.Error("buffered Enter right after the prompt appears must be ignored")
	}
	m.now = func() time.Time { return start.Add(confirmGuard) }
	got, cmd = m.Update(key(tea.KeyEnter))
	if !got.(confirmModel).confirmed || cmd == nil {
		t.Error("Enter after the guard window must confirm")
	}
}

func TestConfirm_EscWorksDuringGuard(t *testing.T) {
	start := time.Now()
	m := newConfirmModel(testEntry)
	m.shownAt = start
	m.now = func() time.Time { return start }
	got, cmd := m.Update(key(tea.KeyEsc))
	if got.(confirmModel).confirmed || cmd == nil {
		t.Error("Esc must cancel even during the guard window")
	}
}

func TestRequireTerminal_RejectsPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if err := requireTerminal(r); err == nil {
		t.Error("requireTerminal must reject a non-terminal stdin")
	}
}
