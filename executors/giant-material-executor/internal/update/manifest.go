package update

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const releaseManifestSchemaVersion = 1

type CodedError struct{ Code string }

func (e *CodedError) Error() string { return e.Code }

func CodeOf(err error) string {
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}

// ReleaseTarget identifies the executable that is allowed to consume a release.
// A Windows archive must never be staged by a macOS executor and vice versa.
type ReleaseTarget struct {
	Platform     string
	Architecture string
}

// ReleaseManifest is the signed contract between a release publisher and one
// executor target. Signature is an Ed25519 signature over SigningBytes().
type ReleaseManifest struct {
	SchemaVersion          int    `json:"schemaVersion"`
	Version                string `json:"version"`
	Platform               string `json:"platform"`
	Architecture           string `json:"architecture"`
	URL                    string `json:"url"`
	SHA256                 string `json:"sha256"`
	MinimumExecutorVersion string `json:"minimumExecutorVersion"`
	PublishedAt            string `json:"publishedAt"`
	Signature              string `json:"signature"`
}

type signedManifest struct {
	SchemaVersion          int    `json:"schemaVersion"`
	Version                string `json:"version"`
	Platform               string `json:"platform"`
	Architecture           string `json:"architecture"`
	URL                    string `json:"url"`
	SHA256                 string `json:"sha256"`
	MinimumExecutorVersion string `json:"minimumExecutorVersion"`
	PublishedAt            string `json:"publishedAt"`
}

// SigningBytes returns the stable, signature-covered representation. The
// signature itself deliberately is not part of the payload.
func (m ReleaseManifest) SigningBytes() ([]byte, error) {
	if err := validateReleaseManifest(m); err != nil {
		return nil, err
	}
	return json.Marshal(signedManifest{
		SchemaVersion:          m.SchemaVersion,
		Version:                strings.TrimSpace(m.Version),
		Platform:               strings.TrimSpace(m.Platform),
		Architecture:           strings.TrimSpace(m.Architecture),
		URL:                    strings.TrimSpace(m.URL),
		SHA256:                 strings.ToLower(strings.TrimSpace(m.SHA256)),
		MinimumExecutorVersion: strings.TrimSpace(m.MinimumExecutorVersion),
		PublishedAt:            strings.TrimSpace(m.PublishedAt),
	})
}

// VerifyReleaseManifest verifies the release's target, compatibility, and
// Ed25519 signature before any archive download is considered.
func VerifyReleaseManifest(manifest ReleaseManifest, publicKey ed25519.PublicKey, target ReleaseTarget, currentVersion string) (ReleaseManifest, error) {
	if err := validateReleaseManifest(manifest); err != nil {
		return ReleaseManifest{}, err
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return ReleaseManifest{}, &CodedError{Code: "UPDATE_MANIFEST_SIGNATURE_INVALID"}
	}
	if strings.TrimSpace(manifest.Platform) != strings.TrimSpace(target.Platform) || strings.TrimSpace(manifest.Architecture) != strings.TrimSpace(target.Architecture) {
		return ReleaseManifest{}, &CodedError{Code: "UPDATE_PLATFORM_MISMATCH"}
	}
	if compareVersions(strings.TrimSpace(currentVersion), strings.TrimSpace(manifest.MinimumExecutorVersion)) < 0 {
		return ReleaseManifest{}, &CodedError{Code: "UPDATE_EXECUTOR_TOO_OLD"}
	}
	if compareVersions(strings.TrimSpace(manifest.Version), strings.TrimSpace(currentVersion)) <= 0 {
		return ReleaseManifest{}, &CodedError{Code: "UPDATE_NOT_AVAILABLE"}
	}
	payload, err := manifest.SigningBytes()
	if err != nil {
		return ReleaseManifest{}, err
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(manifest.Signature))
	if err != nil || !ed25519.Verify(publicKey, payload, signature) {
		return ReleaseManifest{}, &CodedError{Code: "UPDATE_MANIFEST_SIGNATURE_INVALID"}
	}
	return manifest, nil
}

func validateReleaseManifest(manifest ReleaseManifest) error {
	if manifest.SchemaVersion != releaseManifestSchemaVersion || !validVersion(strings.TrimSpace(manifest.Version)) || !validVersion(strings.TrimSpace(manifest.MinimumExecutorVersion)) {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	if !validReleaseTarget(manifest.Platform, manifest.Architecture) {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	parsedURL, err := url.Parse(strings.TrimSpace(manifest.URL))
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	hash := strings.TrimSpace(manifest.SHA256)
	if len(hash) != 64 {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	if _, err := time.Parse(time.RFC3339, strings.TrimSpace(manifest.PublishedAt)); err != nil {
		return &CodedError{Code: "UPDATE_MANIFEST_INVALID"}
	}
	return nil
}

func validReleaseTarget(platform, architecture string) bool {
	switch strings.TrimSpace(platform) {
	case "windows":
		return strings.TrimSpace(architecture) == "amd64"
	case "macos":
		return strings.TrimSpace(architecture) == "universal"
	default:
		return false
	}
}

// PublicKeyFromBase64 decodes the release key embedded at build time. A bad
// key disables self-update rather than accepting an unsigned manifest.
func PublicKeyFromBase64(value string) (ed25519.PublicKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, &CodedError{Code: "UPDATE_MANIFEST_SIGNATURE_INVALID"}
	}
	return ed25519.PublicKey(decoded), nil
}

func validVersion(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" && allNumeric(parts)
}

func allNumeric(parts []string) bool {
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

func compareVersions(left, right string) int {
	if !validVersion(left) || !validVersion(right) {
		return strings.Compare(left, right)
	}
	leftParts, rightParts := strings.Split(left, "."), strings.Split(right, ".")
	for index := range leftParts {
		leftNumber, _ := strconv.Atoi(leftParts[index])
		rightNumber, _ := strconv.Atoi(rightParts[index])
		if leftNumber < rightNumber {
			return -1
		}
		if leftNumber > rightNumber {
			return 1
		}
	}
	return 0
}
