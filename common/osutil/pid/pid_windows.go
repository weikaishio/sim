// +build windows

package pid

import (
	"syscall"
)

// https://msdn.microsoft.com/en-us/library/ms684880(v=vs.85).aspx
// Windows Server 2003 and Windows XP:  This access right is not supported.
const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000

const STILL_ACTIVE = 259

func pidIsExist(pid int) bool {
	p, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	var code uint32
	err = syscall.GetExitCodeProcess(p, &code)
	syscall.Close(p)
	if err != nil {
		return code == STILL_ACTIVE
	}
	return true
}
