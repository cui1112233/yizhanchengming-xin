//go:build darwin

package credential

// The security CLI's interactive `-w` prompt does not accept a pipe when the
// executor is launched as a background process.  It can therefore report a
// successful pairing while persisting an empty Keychain value, which makes the
// next launch unpaired.  A user-private 0600 credential file is predictable
// for both app launches and launchd background runs.
func NewDeviceStore(path string) Store { return NewFileStore(path) }
