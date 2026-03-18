package cmd

import (
	"testing"

	"github.com/VictorRoe/passgen-cli/internal/storage"
)

func TestEnsureIDs_AllPresent(t *testing.T) {
	entries := []storage.Entry{
		{ID: "aaaaaaaa-0000-0000-0000-000000000001"},
		{ID: "aaaaaaaa-0000-0000-0000-000000000002"},
	}
	result, changed := ensureIDs(entries)
	if changed {
		t.Error("expected changed=false when all IDs present")
	}
	if result[0].ID != entries[0].ID || result[1].ID != entries[1].ID {
		t.Error("IDs must not be modified when already set")
	}
}

func TestEnsureIDs_NonePresent(t *testing.T) {
	entries := []storage.Entry{
		{Email: "a@b.com"},
		{Email: "c@d.com"},
	}
	result, changed := ensureIDs(entries)
	if !changed {
		t.Error("expected changed=true when no IDs present")
	}
	for i, e := range result {
		if e.ID == "" {
			t.Errorf("entry[%d] still has empty ID after ensureIDs", i)
		}
		if len(e.ID) != 36 {
			t.Errorf("entry[%d] ID length = %d, want 36 (UUID v4)", i, len(e.ID))
		}
	}
}

func TestEnsureIDs_Mixed(t *testing.T) {
	fixed := "aaaaaaaa-0000-0000-0000-000000000001"
	entries := []storage.Entry{
		{ID: fixed, Email: "a@b.com"},
		{Email: "c@d.com"}, // no ID
	}
	result, changed := ensureIDs(entries)
	if !changed {
		t.Error("expected changed=true when some IDs missing")
	}
	if result[0].ID != fixed {
		t.Errorf("existing ID must not be modified: got %q", result[0].ID)
	}
	if result[1].ID == "" {
		t.Error("missing ID must be filled")
	}
}

func TestMatchEntry_TooShort(t *testing.T) {
	entries := []storage.Entry{{ID: "abcdefgh-0000-0000-0000-000000000001"}}
	_, _, err := matchEntry(entries, "abcde") // 5 chars < 8
	if err == nil {
		t.Error("expected error for prefix shorter than 8 chars")
	}
}

func TestMatchEntry_NoMatch(t *testing.T) {
	entries := []storage.Entry{{ID: "abcdefgh-0000-0000-0000-000000000001"}}
	_, _, err := matchEntry(entries, "zzzzzzzzz")
	if err == nil {
		t.Error("expected error for no match")
	}
}

func TestMatchEntry_Ambiguous(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdefgh-0000-0000-0000-000000000001"},
		{ID: "abcdefgh-0000-0000-0000-000000000002"},
	}
	_, _, err := matchEntry(entries, "abcdefgh")
	if err == nil {
		t.Error("expected error for ambiguous prefix")
	}
}

func TestMatchEntry_ExactOne(t *testing.T) {
	entries := []storage.Entry{
		{ID: "abcdefgh-0000-0000-0000-000000000001", Service: "google"},
		{ID: "xxxxxxxx-0000-0000-0000-000000000002", Service: "github"},
	}
	entry, idx, err := matchEntry(entries, "abcdefgh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Service != "google" {
		t.Errorf("wrong entry returned: got %q", entry.Service)
	}
	if idx != 0 {
		t.Errorf("wrong index: got %d want 0", idx)
	}
}
