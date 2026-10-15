//go:build windows

package credential

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protect(data []byte) ([]byte, error) {
	return cryptProtect(data, windows.CryptProtectData)
}

func unprotect(data []byte) ([]byte, error) {
	return cryptUnprotect(data, windows.CryptUnprotectData)
}

func cryptProtect(data []byte, fn func(*windows.DataBlob, *uint16, *windows.DataBlob, uintptr, *windows.CryptProtectPromptStruct, uint32, *windows.DataBlob) error) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty credential payload")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := fn(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	return copyAndFree(&out)
}

func cryptUnprotect(data []byte, fn func(*windows.DataBlob, **uint16, *windows.DataBlob, uintptr, *windows.CryptProtectPromptStruct, uint32, *windows.DataBlob) error) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty credential payload")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := fn(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return nil, err
	}
	return copyAndFree(&out)
}

func copyAndFree(blob *windows.DataBlob) ([]byte, error) {
	if blob == nil || blob.Data == nil || blob.Size == 0 {
		return nil, errors.New("empty protected credential payload")
	}
	data := append([]byte(nil), unsafe.Slice(blob.Data, blob.Size)...)
	_, _ = windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(blob.Data))))
	return data, nil
}
