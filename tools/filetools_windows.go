// +build windows

package tools

import (
	"golang.org/x/sys/windows"
)

func openSymlink(path string) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	attrs := uint32(windows.FILE_FLAG_BACKUP_SEMANTICS)
	h, err := windows.CreateFile(p, 0, 0, nil, windows.OPEN_EXISTING, attrs, 0)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func CanonicalizeSystemPath(path string) (string, error) {
	h, err := openSymlink(path)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	buf := make([]uint16, 100)
	for {
		n, err := windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), 0)
		if err != nil {
			return "", err
		}
		if n < uint32(len(buf)) {
			break
		}
		buf = make([]uint16, n)
	}

	s := windows.UTF16ToString(buf)
	if len(s) > 4 && s[:4] == `\\?\` {
		s = s[4:]
		if len(s) > 3 && s[:3] == `UNC` {
			// return path like \\server\share\...
			return `\` + s[3:], nil
		}
		return s, nil
	}
	return s, nil
}
