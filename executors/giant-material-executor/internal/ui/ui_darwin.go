//go:build darwin

package ui

import (
	"context"
	"log"
	"os/exec"
)

func Run(ctx context.Context, config Config) error {
	if err := validateConfig(config); err != nil {
		return err
	}
	const setupURL = "http://127.0.0.1:17861/setup"
	if err := exec.Command("/usr/bin/open", setupURL).Start(); err != nil {
		log.Printf("open setup page manually: %s (%v)", setupURL, err)
	}
	<-ctx.Done()
	return nil
}
