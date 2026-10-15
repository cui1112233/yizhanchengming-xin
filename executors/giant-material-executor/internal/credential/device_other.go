//go:build !darwin

package credential

func NewDeviceStore(path string) Store { return NewProtectedStore(path) }
