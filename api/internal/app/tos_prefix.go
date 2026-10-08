package app

import (
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
)

// configureTOSKeyPrefix gates every TOS-backed dependency on one validated
// namespace. Invalid nonblank configuration must not fall back to production.
func configureTOSKeyPrefix(raw string, wire func(objectkey.Prefix)) error {
	prefix, err := objectkey.ParsePrefix(raw)
	if err != nil {
		return errors.New("app: TOS key prefix configuration is invalid")
	}
	wire(prefix)
	return nil
}
