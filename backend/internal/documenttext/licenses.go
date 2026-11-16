package documenttext

import (
	"embed"
	"io/fs"
	"strings"
)

// Licenses is included in the deployed single binary alongside the WASM engine.
//
//go:embed licenses
var Licenses embed.FS

func LicenseText() string {
	var b strings.Builder
	fs.WalkDir(Licenses, "licenses", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, e := Licenses.ReadFile(path)
		if e != nil {
			return e
		}
		b.WriteString("\n=== " + path + " ===\n")
		b.Write(raw)
		return nil
	})
	return b.String()
}
