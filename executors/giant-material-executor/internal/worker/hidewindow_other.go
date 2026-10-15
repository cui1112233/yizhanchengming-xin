//go:build !windows

package worker

import "os/exec"

func hideConsoleWindow(*exec.Cmd) {}
