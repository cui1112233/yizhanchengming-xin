package generation

import "testing"

func TestSecurityAudioAssetRejectsTraversalAndAbsolutePaths(t *testing.T) {
	for _, asset := range []string{
		"../../etc/passwd",
		"../private/audio.wav",
		"/etc/passwd",
		`C:\\Windows\\System32\\drivers\\etc\\hosts`,
	} {
		if safeLocalAudioAsset(asset) {
			t.Fatalf("safeLocalAudioAsset(%q) = true, want false for browser-controlled traversal/absolute path", asset)
		}
	}
}

func TestSecurityAudioAssetAllowsServerRelativeAssetPath(t *testing.T) {
	if !safeLocalAudioAsset("runtime/audio/project-51/book-7.wav") {
		t.Fatal("expected normalized server-relative audio asset path to remain allowed")
	}
}
