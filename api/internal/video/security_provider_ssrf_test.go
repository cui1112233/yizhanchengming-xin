package video

import "testing"

func TestSecurityProviderEndpointBlocksMetadataAndPrivateHTTPS(t *testing.T) {
	for _, raw := range []string{
		"https://169.254.169.254/latest/meta-data/",
		"https://10.0.0.8/api/video",
		"https://192.168.1.8/api/video",
	} {
		if err := validateProviderURL(raw); err == nil {
			t.Fatalf("validateProviderURL(%q) = nil, want blocked private/link-local endpoint", raw)
		}
	}
}

func TestSecurityProviderEndpointKeepsExplicitLocalDevelopmentLoopback(t *testing.T) {
	for _, raw := range []string{
		"http://127.0.0.1:8188/api/video",
		"http://localhost:8188/api/video",
	} {
		if err := validateProviderURL(raw); err != nil {
			t.Fatalf("validateProviderURL(%q) = %v, want explicit local loopback allowed", raw, err)
		}
	}
}
