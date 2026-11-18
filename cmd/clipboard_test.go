package cmd

import (
	"errors"
	"syscall"
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

func TestNormalizeRead_EmptyClipboardIsNotAnError(t *testing.T) {
	// On Windows atotto returns Errno(0) ("operation completed successfully")
	// when the clipboard holds no text.
	s, err := normalizeRead("", syscall.Errno(0))
	if err != nil || s != "" {
		t.Errorf("normalizeRead = (%q, %v), want (\"\", nil)", s, err)
	}
}

func TestNormalizeRead_RealErrorKept(t *testing.T) {
	boom := errors.New("boom")
	if _, err := normalizeRead("", boom); !errors.Is(err, boom) {
		t.Errorf("normalizeRead must keep real errors, got %v", err)
	}
}
