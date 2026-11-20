package novelcreation

// The exact key includes the chapter/piece/chunk suffix. A retry must never
// discard successful sibling calls simply because they use the same prompt.
type shortCallError struct {
	key   string
	cause error
}

func (e *shortCallError) Error() string { return e.cause.Error() }
func (e *shortCallError) Unwrap() error { return e.cause }
func shortCallFailure(in RunContext, key string, e error) error {
	if in.ShortImitationVersion >= 1 && IsCode(e, ErrInvalidOutput) {
		return &shortCallError{key, e}
	}
	return e
}
