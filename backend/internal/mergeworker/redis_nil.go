package mergeworker

import "errors"

// IsRedisNil distinguishes an absent Redis value from a connection failure.
func IsRedisNil(err error) bool { return errors.Is(err, errRedisNil) }
