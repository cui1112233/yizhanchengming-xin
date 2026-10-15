//go:build !darwin && !windows

package main

import "fmt"

func applyUpdate(request updateRequest) error {
	return fmt.Errorf("self update is unavailable on %s", request.Platform)
}
