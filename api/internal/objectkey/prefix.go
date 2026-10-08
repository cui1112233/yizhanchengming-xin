// Package objectkey normalizes namespaces for newly written storage objects.
// Persisted keys are complete references and must never be prefixed again.
package objectkey

import (
	"errors"
	"strings"
)

// Prefix is an immutable normalized prefix. Its zero value preserves keys.
type Prefix struct{ value string }

func ParsePrefix(raw string) (Prefix, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return Prefix{}, nil
	}
	value = strings.TrimRight(value, "/")
	if !validRelativeKey(value) {
		return Prefix{}, errors.New("objectkey: invalid prefix")
	}
	return Prefix{value: value + "/"}, nil
}

// Apply validates a new relative key before adding the namespace once.
// Empty-prefix mode intentionally preserves legacy key behavior.
func (p Prefix) Apply(relativeKey string) (string, error) {
	if p.value == "" {
		return relativeKey, nil
	}
	if !validRelativeKey(relativeKey) || strings.HasPrefix(relativeKey, p.value) {
		return "", errors.New("objectkey: invalid relative key")
	}
	return p.value + relativeKey, nil
}

func validRelativeKey(key string) bool {
	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
		for _, char := range segment {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '.' || char == '_' || char == '-') {
				return false
			}
		}
	}
	return true
}
