package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

// RegisterApplicationsCompat replaces the small public account-application
// API that used to live only in Express. It deliberately shares the same
// applications.json file consumed by the Go admin review handlers.
func RegisterApplicationsCompat(root *http.ServeMux, accounts auth.Store, profiles *MemberProfileStore, now func() time.Time) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	root.Handle("POST /api/applications", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Reason   string `json:"reason"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&input) != nil ||
			!managedUsernamePattern.MatchString(strings.TrimSpace(input.Username)) || len(input.Password) < 8 || strings.TrimSpace(input.Reason) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
			return
		}
		username := strings.TrimSpace(input.Username)
		if _, err := accounts.FindAccount(req.Context(), username); err == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Username is unavailable"})
			return
		} else if err != auth.ErrNotFound {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		created := now().UTC().Format(time.RFC3339Nano)
		id := applicationID(now)
		profiles.mu.Lock()
		defer profiles.mu.Unlock()
		rows := readAdminGovernancePrivateArray(adminSystemFile(profiles.SystemDir, "applications.json"))
		for _, row := range rows {
			if strings.EqualFold(toString(row["username"]), username) && strings.EqualFold(toString(row["status"]), "pending") {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "Username is unavailable"})
				return
			}
		}
		row := map[string]any{"id": id, "username": username, "passwordHash": hash, "passwordAlgorithm": algorithm, "reason": strings.TrimSpace(input.Reason), "status": "pending", "createdAt": created, "updatedAt": created, "reviewedAt": nil, "reviewedBy": nil}
		rows = append(rows, row)
		if err := writeAdminGovernanceArray(profiles, "applications.json", rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, publicApplicationCompat(row))
	}))

	root.Handle("POST /api/applications/{id}/status", applicationStatusHandler(accounts, profiles))
	root.Handle("POST /api/applications/{id}/withdraw", applicationWithdrawHandler(accounts, profiles, now))
}

func applicationStatusHandler(accounts auth.Store, profiles *MemberProfileStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct{ Username, Password string }
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
			return
		}
		row, ok := findApplicationCompat(profiles, req.PathValue("id"))
		if !ok || toString(row["username"]) != strings.TrimSpace(input.Username) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		account, err := accounts.FindAccount(req.Context(), input.Username)
		if err == nil && auth.VerifyPassword(input.Password, account.PasswordHash, account.PasswordAlgorithm) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		if !auth.VerifyPassword(input.Password, toString(row["passwordHash"]), toString(row["passwordAlgorithm"])) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		writeJSON(w, http.StatusOK, publicApplicationCompat(row))
	})
}

func applicationWithdrawHandler(accounts auth.Store, profiles *MemberProfileStore, now func() time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct{ Username, Password string }
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
			return
		}
		profiles.mu.Lock()
		defer profiles.mu.Unlock()
		rows := readAdminGovernancePrivateArray(adminSystemFile(profiles.SystemDir, "applications.json"))
		index := -1
		for i, row := range rows {
			if toString(row["id"]) == req.PathValue("id") {
				index = i
				break
			}
		}
		if index < 0 || toString(rows[index]["username"]) != strings.TrimSpace(input.Username) || !auth.VerifyPassword(input.Password, toString(rows[index]["passwordHash"]), toString(rows[index]["passwordAlgorithm"])) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Application not found"})
			return
		}
		if toString(rows[index]["status"]) != "pending" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Application is not pending"})
			return
		}
		rows[index]["status"] = "withdrawn"
		rows[index]["updatedAt"] = now().UTC().Format(time.RFC3339Nano)
		if err := writeAdminGovernanceArray(profiles, "applications.json", rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, publicApplicationCompat(rows[index]))
	})
}

func findApplicationCompat(profiles *MemberProfileStore, id string) (map[string]any, bool) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	for _, row := range readAdminGovernancePrivateArray(adminSystemFile(profiles.SystemDir, "applications.json")) {
		if toString(row["id"]) == id {
			return row, true
		}
	}
	return nil, false
}

func publicApplicationCompat(row map[string]any) map[string]any {
	public := cloneAdminGovernanceMap(row)
	delete(public, "passwordHash")
	delete(public, "passwordAlgorithm")
	return public
}

func applicationID(now func() time.Time) string {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "app-" + now().UTC().Format("20060102150405.000000000")
	}
	return "app-" + now().UTC().Format("20060102150405") + "-" + hex.EncodeToString(raw)
}
