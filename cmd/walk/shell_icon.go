package main

import (
	"log"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

func getShellIcon(path string) *walk.Icon {
	var shfi win.SHFILEINFO
	ret := win.SHGetFileInfo(
		syscall.StringToUTF16Ptr(path),
		0,
		&shfi,
		uint32(unsafe.Sizeof(shfi)),
		win.SHGFI_ICON|win.SHGFI_SMALLICON,
	)
	if ret == 0 {
		log.Printf("SHGetFileInfo falhou para %s", path)
		return nil
	}
	icon, err := walk.NewIconFromHICON(shfi.HIcon)
	if err != nil {
		log.Printf("NewIconFromHICON falhou: %v", err)
		win.DestroyIcon(shfi.HIcon)
		return nil
	}
	return icon
}
