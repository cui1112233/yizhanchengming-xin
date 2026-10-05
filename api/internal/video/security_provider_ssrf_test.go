package video

import "testing"

func TestSecurityProviderEndpointBlocksMetadataAndPrivateHTTPS(t *testing.T) {
	for _, raw := range []string{
		"https://169.254.169.254/latest/meta-data/",
		"https://10.0.0.8/api/video",
		"https://172.16.0.8/api/video",
		"https://192.168.1.8/api/video",
		"https://[::1]/api/video",
		"https://[fc00::1]/api/video",
		"https://[fe80::1]/api/video",
		"https://[::ffff:127.0.0.1]/api/video",
	} {
		if err := validateProviderURL(raw); err == nil {
			t.Fatalf("validateProviderURL(%q) = nil, want blocked private/link-local endpoint", raw)
		}
	}
}

func TestSecurityProviderEndpointBlocksObfuscatedNumericLoopback(t *testing.T) {
	for _, raw := range []string{
		"https://2130706433/api/video",
		"https://0x7f000001/api/video",
		"https://0177.0.0.1/api/video",
		"https://0x7f.0x0.0x0.0x1/api/video",
	} {
		if err := validateProviderURL(raw); err == nil {
			t.Fatalf("validateProviderURL(%q) = nil, want obfuscated numeric host blocked before outbound resolution", raw)
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
