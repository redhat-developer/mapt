// Package signalwindows sends a control break to another process.

//go:build windows
// +build windows

package sigintwindows

import (
	"syscall"
)

// SendCtrlBreak sends a Ctrl-Break event to the process with id pid
func SendCtrlBreak(pid int) error {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return e
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return e
	}
	r, _, e := p.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(pid))
	if r == 0 {
		return e // syscall.GetLastError()
	}
	return nil
}
