package httpapi

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type novelPanelLease struct {
	instanceID string
	expiresAt  int64
}

var novelPanelLeases = struct {
	sync.Mutex
	items map[string]novelPanelLease
}{items: make(map[string]novelPanelLease)}

func RegisterNovelPanelLeaseCompat(root *http.ServeMux, middleware BridgeAuth) {
	root.Handle("POST /api/novel-panel/character-core/project-lease", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "登录已失效，请重新登录"})
			return
		}
		var input struct {
			Action     string `json:"action"`
			ProjectID  string `json:"project_id"`
			InstanceID string `json:"instance_id"`
			Force      bool   `json:"force"`
		}
		if !decodeJSON(w, req, &input) {
			return
		}
		input.Action = strings.TrimSpace(input.Action)
		input.ProjectID = strings.TrimSpace(input.ProjectID)
		input.InstanceID = strings.TrimSpace(input.InstanceID)
		if input.Action != "acquire" && input.Action != "heartbeat" && input.Action != "release" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid lease action"})
			return
		}
		if input.ProjectID == "" || len(input.ProjectID) > 80 || input.InstanceID == "" || len(input.InstanceID) > 256 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid project lease"})
			return
		}

		now := time.Now().UnixMilli()
		key := identity.Username + "\x00" + input.ProjectID
		novelPanelLeases.Lock()
		defer novelPanelLeases.Unlock()
		current, exists := novelPanelLeases.items[key]
		if exists && current.expiresAt <= now {
			delete(novelPanelLeases.items, key)
			exists = false
		}
		response := map[string]any{"acquired": false, "released": false, "holder_instance_id": nil, "expires_at": nil}
		if exists {
			response["holder_instance_id"] = current.instanceID
			response["expires_at"] = current.expiresAt
		}
		if input.Action == "release" {
			if !exists || current.instanceID != input.InstanceID {
				writeJSON(w, http.StatusOK, response)
				return
			}
			delete(novelPanelLeases.items, key)
			response["released"] = true
			response["holder_instance_id"] = nil
			response["expires_at"] = nil
			writeJSON(w, http.StatusOK, response)
			return
		}
		if input.Action == "heartbeat" && (!exists || current.instanceID != input.InstanceID) {
			writeJSON(w, http.StatusOK, response)
			return
		}
		if exists && current.instanceID != input.InstanceID && !input.Force {
			writeJSON(w, http.StatusOK, response)
			return
		}
		next := novelPanelLease{instanceID: input.InstanceID, expiresAt: now + 90_000}
		novelPanelLeases.items[key] = next
		response["acquired"] = true
		response["holder_instance_id"] = next.instanceID
		response["expires_at"] = next.expiresAt
		writeJSON(w, http.StatusOK, response)
	})))
}
