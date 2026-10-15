//go:build !windows && !darwin

package ui

import "context"

func Run(_ context.Context, config Config) error {
	return validateConfig(config)
}
