package video

import "strings"

func safeExecutorErrorMessage(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	lower := strings.ToLower(message)
	for _, marker := range []string{
		"authorization", "bearer ", "cookie", "password", "passwd", "token",
		"api key", "api_key", "apikey", "secret", "credential", "ciphertext",
		"nonce", "master key", "mysql://", "dsn", "/srv/", "/home/", "/etc/",
		"c:\\", "d:\\",
	} {
		if strings.Contains(lower, marker) {
			return "local executor failed; sensitive details redacted"
		}
	}
	if len(message) > 240 {
		return message[:240]
	}
	return message
}
