package modelcache

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

type Manifest struct {
	Version        string `json:"version"`
	PackageURL     string `json:"packageUrl"`
	Size           int64  `json:"size"`
	SHA256         string `json:"sha256"`
	RuntimeVersion string `json:"runtimeVersion"`
}

func (m Manifest) Validate() error {
	if strings.TrimSpace(m.Version) == "" || m.Version == "." || m.Version == ".." || strings.ContainsAny(m.Version, `/\\`) {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: unsafe version")
	}
	if m.Size <= 0 {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: size must be positive")
	}
	if strings.TrimSpace(m.RuntimeVersion) == "" {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: runtimeVersion is required")
	}
	parsed, err := url.Parse(m.PackageURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: packageUrl must be http(s)")
	}
	sha := strings.TrimSpace(m.SHA256)
	if len(sha) != 64 {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: sha256 must be 64 hex characters")
	}
	if _, err := hex.DecodeString(sha); err != nil {
		return fmt.Errorf("MODEL_MANIFEST_INVALID: sha256 is not hexadecimal")
	}
	return nil
}
