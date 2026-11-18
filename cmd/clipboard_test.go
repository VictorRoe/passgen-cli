package cmd

import (
	"errors"
	"testing"
)

type fakeClipboard struct {
	content string
	readErr error
	writes  int
}

func (f *fakeClipboard) read() (string, error) { return f.content, f.readErr }
func (f *fakeClipboard) write(s string) error  { f.content = s; f.writes++; return nil }

func TestClearIfUnchanged_Clears(t *testing.T) {
	c := &fakeClipboard{content: "secret"}
	cleared, err := clearIfUnchanged(c.read, c.write, "secret")
	if err != nil || !cleared || c.content != "" {
		t.Errorf("cleared=%v err=%v content=%q, want true nil \"\"", cleared, err, c.content)
	}
}

func TestClearIfUnchanged_KeepsOther(t *testing.T) {
	c := &fakeClipboard{content: "something else"}
	cleared, err := clearIfUnchanged(c.read, c.write, "secret")
	if err != nil || cleared || c.writes != 0 {
		t.Errorf("cleared=%v err=%v writes=%d, want false nil 0", cleared, err, c.writes)
	}
}

func TestClearIfUnchanged_ReadError(t *testing.T) {
	c := &fakeClipboard{readErr: errors.New("boom")}
	cleared, err := clearIfUnchanged(c.read, c.write, "secret")
	if err == nil || cleared || c.writes != 0 {
		t.Errorf("cleared=%v err=%v writes=%d, want false error 0", cleared, err, c.writes)
	}
}
