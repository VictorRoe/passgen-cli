//go:build windows

package session

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// secureDir replaces dir's ACL with a single, protected (non-inherited) entry
// granting full control to the current user only. Child files and the
// socket inherit it. Without this, only the profile folder's default
// permissions would protect the session socket.
func secureDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("session: crear directorio: %w", err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("session: obtener usuario actual: %w", err)
	}
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid),
		},
	}}, nil)
	if err != nil {
		return fmt.Errorf("session: crear ACL: %w", err)
	}
	err = windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil)
	if err != nil {
		return fmt.Errorf("session: aplicar ACL a %s: %w", dir, err)
	}
	return nil
}
