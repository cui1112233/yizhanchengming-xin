package httpapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var mfaCodePattern = regexp.MustCompile(`^\d{6}$`)

type mfaFileRow struct {
	Username       string   `json:"username"`
	Enabled        bool     `json:"enabled"`
	Secret         string   `json:"secret"`
	RecoveryHashes []string `json:"recoveryHashes"`
	PendingSecret  string   `json:"pendingSecret"`
}

func readMFAFile(path string) ([]mfaFileRow, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rows []mfaFileRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("读取 MFA 配置失败")
	}
	return rows, nil
}

func normalizeMFARecoveryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func mfaRecoveryHash(value string) string {
	sum := sha256.Sum256([]byte(normalizeMFARecoveryCode(value)))
	return hex.EncodeToString(sum[:])
}

func decodeMFASecret(value string) ([]byte, error) {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), "=", ""))
	clean = strings.ReplaceAll(clean, " ", "")
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(clean)
}

func mfaTOTP(secret, code string, now time.Time) bool {
	if !mfaCodePattern.MatchString(strings.TrimSpace(code)) {
		return false
	}
	key, err := decodeMFASecret(secret)
	if err != nil {
		return false
	}
	counter := now.Unix() / 30
	for drift := int64(-1); drift <= 1; drift++ {
		var message [8]byte
		binary.BigEndian.PutUint64(message[:], uint64(counter+drift))
		h := hmac.New(sha1.New, key)
		_, _ = h.Write(message[:])
		sum := h.Sum(nil)
		offset := sum[len(sum)-1] & 0x0f
		value := (uint32(sum[offset])&0x7f)<<24 | uint32(sum[offset+1])<<16 | uint32(sum[offset+2])<<8 | uint32(sum[offset+3])
		if fmt.Sprintf("%06d", value%1_000_000) == strings.TrimSpace(code) {
			return true
		}
	}
	return false
}

func verifyMFAForLogin(path, username, code string, now time.Time) (required, valid, recoveryUsed bool, err error) {
	rows, err := readMFAFile(path)
	if err != nil {
		return false, false, false, err
	}
	for index := range rows {
		row := &rows[index]
		if row.Username != username || !row.Enabled || strings.TrimSpace(row.Secret) == "" {
			continue
		}
		required = true
		if mfaTOTP(row.Secret, code, now) {
			return true, true, false, nil
		}
		hash := mfaRecoveryHash(code)
		for recoveryIndex, candidate := range row.RecoveryHashes {
			if candidate != hash {
				continue
			}
			row.RecoveryHashes = append(row.RecoveryHashes[:recoveryIndex], row.RecoveryHashes[recoveryIndex+1:]...)
			data, marshalErr := json.MarshalIndent(rows, "", "  ")
			if marshalErr != nil {
				return true, false, false, marshalErr
			}
			if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
				return true, false, false, writeErr
			}
			return true, true, true, nil
		}
		return true, false, false, nil
	}
	return false, true, false, nil
}

func writeMFAFile(path string, rows []mfaFileRow) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func base32MFASecret() (string, error) {
	seed := make([]byte, 20)
	if _, err := rand.Read(seed); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(seed), nil
}

func generateMFARecoveryCodes() ([]string, error) {
	result := make([]string, 8)
	for index := range result {
		seed := make([]byte, 5)
		if _, err := rand.Read(seed); err != nil {
			return nil, err
		}
		hexValue := strings.ToUpper(hex.EncodeToString(seed))
		result[index] = hexValue[:5] + "-" + hexValue[5:]
	}
	return result, nil
}

func mfaStatusRow(row *mfaFileRow) map[string]any {
	status := map[string]any{"enabled": false, "enabledAt": nil, "recoveryCodesRemaining": 0, "setupPending": false}
	if row == nil {
		return status
	}
	status["enabled"] = row.Enabled && strings.TrimSpace(row.Secret) != ""
	status["recoveryCodesRemaining"] = len(row.RecoveryHashes)
	status["setupPending"] = strings.TrimSpace(row.PendingSecret) != ""
	return status
}
