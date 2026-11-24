//go:build windows

package fs

import (
	"os"
	"runtime"
	"unsafe"

	"github.com/ygrebnov/errorc"
	"github.com/ygrebnov/keys"
	"golang.org/x/sys/windows"
)

var (
	keyCause   = keys.New("cause")
	keySID     = keys.New("sid")
	keyACEType = keys.New("ace_type")
)

func validatePrivatePermissions(file *os.File) error {
	defer runtime.KeepAlive(file)
	return validatePrivateACL(windows.Handle(file.Fd()))
}

func setPrivateDirectory(dir *os.File) error {
	defer runtime.KeepAlive(dir)
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}

	sddl := "D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}

	acl, _, err := descriptor.DACL()
	if err != nil {
		return err
	}

	// os.Open does not request WRITE_DAC, so its handle cannot update the ACL.
	// Reopen the directory with ACL read and write rights, then verify that
	// both handles refer to the same object before changing its descriptor.
	original := windows.Handle(dir.Fd())
	path, err := finalPathName(original)
	if err != nil {
		return err
	}

	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	writable, err := windows.CreateFile(
		name, windows.READ_CONTROL|windows.WRITE_DAC|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(writable)

	var originalInfo, writableInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(original, &originalInfo); err != nil {
		return err
	}

	if err := windows.GetFileInformationByHandle(writable, &writableInfo); err != nil {
		return err
	}

	if originalInfo.VolumeSerialNumber != writableInfo.VolumeSerialNumber ||
		originalInfo.FileIndexHigh != writableInfo.FileIndexHigh ||
		originalInfo.FileIndexLow != writableInfo.FileIndexLow {
		return ErrDirectoryChanged
	}

	return windows.SetSecurityInfo(
		writable, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, acl, nil,
	)
}

func finalPathName(handle windows.Handle) (string, error) {
	// FILE_NAME_OPENED avoids normalizing every path component, which can
	// require access that the caller does not have. Handle identity is checked
	// after reopening, so the returned name is only a candidate path.
	const fileNameOpened = 0x8
	size, err := windows.GetFinalPathNameByHandle(handle, nil, 0, fileNameOpened)
	if err != nil {
		return "", err
	}
	for {
		buf := make([]uint16, size+1)
		n, err := windows.GetFinalPathNameByHandle(handle, &buf[0], uint32(len(buf)), fileNameOpened)
		if err != nil {
			return "", err
		}
		if n < uint32(len(buf)) {
			return windows.UTF16ToString(buf[:n]), nil
		}
		size = n
	}
}

func validatePrivateACL(handle windows.Handle) error {
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}

	if sd == nil {
		return errorc.With(
			ErrInsecurePermissions,
			errorc.String(keyCause, "missing security descriptor"),
		)
	}

	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	if dacl == nil {
		return errorc.With(
			ErrInsecurePermissions,
			errorc.String(keyCause, "NULL DACL"),
		)
	}

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}

	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}

	admins, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return err
	}

	allowed := []*windows.SID{user.User.Sid, system, admins}

	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if errACE := windows.GetAce(dacl, i, &ace); errACE != nil {
			return errACE
		}

		switch ace.Header.AceType {
		case windows.ACCESS_DENIED_ACE_TYPE:
			continue
		case windows.ACCESS_ALLOWED_ACE_TYPE:
			if ace.Header.AceFlags&windows.INHERIT_ONLY_ACE != 0 {
				continue
			}

			sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
			if !sid.IsValid() {
				return errorc.With(
					ErrInsecurePermissions,
					errorc.String(keyCause, "invalid SID"),
				)
			}

			permitted := false
			for _, candidate := range allowed {
				if sid.Equals(candidate) {
					permitted = true
					break
				}
			}
			if !permitted {
				return errorc.With(
					ErrInsecurePermissions,
					errorc.String(keyCause, "ACL grants access to SID"),
					errorc.String(keySID, sid.String()),
				)
			}
		default:
			return errorc.With(
				ErrInsecurePermissions,
				errorc.String(keyCause, "unsupported ACE type"),
				errorc.Int(keyACEType, int(ace.Header.AceType)),
			)
		}
	}

	return nil
}
