//go:build windows

package shell

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SHFILEOPSTRUCT é a estrutura passada para SHFileOperation.
// Os caminhos em pFrom e pTo são terminados com duplo-nulo.
type shFileOpStruct struct {
	Hwnd                  uintptr
	WFunc                 uint32
	PFrom                 *uint16
	PTo                   *uint16
	FFlags                uint16
	FAnyOperationsAborted int32
	HNameMappings         uintptr
	LpszProgressTitle     *uint16
}

const (
	foDelete          = 0x0003
	fofAllowUndo      = 0x0040 // manda para a lixeira
	fofNoConfirmation = 0x0010 // não pede confirmação
	fofSilent         = 0x0004 // sem dialog de progresso do shell (usamos o nosso)
	fofNoErrorUI      = 0x0400 // sem dialog de erro do shell
)

var (
	shell32         = windows.NewLazySystemDLL("shell32.dll")
	shFileOperation = shell32.NewProc("SHFileOperationW")
)

// RecycleItems move os caminhos fornecidos para a Lixeira do Windows.
// Retorna erro se a operação falhar.
func RecycleItems(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	from := multiStringToDoubleNull(paths)

	op := shFileOpStruct{
		Hwnd:   0,
		WFunc:  foDelete,
		PFrom:  &from[0],
		PTo:    nil,
		FFlags: fofAllowUndo | fofNoConfirmation | fofNoErrorUI,
	}

	ret, _, _ := shFileOperation.Call(uintptr(unsafe.Pointer(&op)))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

// multiStringToDoubleNull converte []string em buffer UTF-16 com double-null terminator.
func multiStringToDoubleNull(paths []string) []uint16 {
	var buf []uint16
	for _, p := range paths {
		encoded, err := syscall.UTF16FromString(p)
		if err != nil {
			continue
		}
		buf = append(buf, encoded...) // já inclui nulo terminal de cada string
	}
	buf = append(buf, 0) // nulo extra = double-null terminator
	return buf
}
