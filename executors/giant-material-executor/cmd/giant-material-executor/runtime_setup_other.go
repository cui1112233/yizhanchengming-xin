//go:build !windows

package main

import "context"

func ensureRuntime(_ context.Context, _, _ string) error { return nil }
