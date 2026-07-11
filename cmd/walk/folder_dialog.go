//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modShell32               = windows.NewLazySystemDLL("shell32.dll")
	procSHBrowseForFolderW   = modShell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = modShell32.NewProc("SHGetPathFromIDListW")
)

type browseInfoW struct {
	HwndOwner      uintptr
	PIDLRoot       uintptr
	PSZDisplayName uintptr
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}

const bifReturnOnlyFSDirs = 0x0001
const bifNewDialogStyle = 0x0040

func showFolderDialog(owner *GarqMainWindow, title, initialPath string) string {
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	displayBuf := make([]uint16, 260)

	bi := &browseInfoW{
		HwndOwner:      uintptr(unsafe.Pointer(owner.MainWindow.Handle())),
		PSZDisplayName: uintptr(unsafe.Pointer(&displayBuf[0])),
		LpszTitle:      titlePtr,
		UlFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
	}

	ret, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(bi)))
	if ret == 0 {
		return ""
	}

	pathBuf := make([]uint16, 260)
	procSHGetPathFromIDListW.Call(ret, uintptr(unsafe.Pointer(&pathBuf[0])))
	return syscall.UTF16ToString(pathBuf)
}
