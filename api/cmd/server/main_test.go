package main

import "testing"

func TestBuildInfoUsesLinkerInjectedGitSHA(t *testing.T) {
	previous := BuildGitSHA
	BuildGitSHA = "release-sha-123"
	t.Cleanup(func() { BuildGitSHA = previous })

	if got := buildInfo().GitSHA; got != "release-sha-123" {
		t.Fatalf("gitSha=%q want release-sha-123", got)
	}
}
