//go:build windows

package session

import (
	"os"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestSecureDir_WindowsOwnerOnly(t *testing.T) {
	dir, err := os.MkdirTemp("", "pg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := secureDir(dir); err != nil {
		t.Fatalf("secureDir() error = %v", err)
	}

	sd, err := windows.GetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := sd.Control()
	if err != nil {
		t.Fatal(err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Error("DACL must be protected (no inherited entries)")
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if acl.AceCount == 0 {
		t.Fatal("DACL has no entries")
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	// Windows stores an inheritable grant as two entries (this folder +
	// inherit-only); every entry must belong to the current user.
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			t.Fatal(err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			t.Errorf("entry %d type = %d, want ACCESS_ALLOWED", i, ace.Header.AceType)
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.Equals(user.User.Sid) {
			t.Errorf("entry %d SID = %s, want current user %s", i, sid, user.User.Sid)
		}
	}
}
