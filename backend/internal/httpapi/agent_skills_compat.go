package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var agentSkillIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func RegisterAgentSkillsCompat(root *http.ServeMux, middleware BridgeAuth, systemDir, usersDir string) {
	if root == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	root.Handle("GET /api/agent/skills", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		skills := []any{}
		for _, skill := range readAgentSkills(filepath.Join(systemDir, "agent-skills.json")) {
			if toString(skill["source"]) == "system" && toString(skill["status"]) == "published" {
				skills = append(skills, publicAgentSkill(skill))
			}
		}
		for _, skill := range readAgentSkills(agentSkillsPath(usersDir, identity.Username)) {
			if toString(skill["source"]) == "user" && toString(skill["owner"]) == identity.Username && toString(skill["status"]) != "archived" {
				skills = append(skills, publicAgentSkill(skill))
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"skills": skills})
	})))
	root.Handle("GET /api/agent/skills/mine/{id}", middleware.Middleware(agentSkillMineGet(usersDir)))
	root.Handle("POST /api/agent/skills/mine", middleware.Middleware(agentSkillCreate(usersDir)))
	root.Handle("PUT /api/agent/skills/mine/{id}", middleware.Middleware(agentSkillUpdate(usersDir)))
	root.Handle("DELETE /api/agent/skills/mine/{id}", middleware.Middleware(agentSkillDelete(usersDir)))
}

func agentSkillMineGet(usersDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := req.PathValue("id")
		for _, skill := range readAgentSkills(agentSkillsPath(usersDir, identity.Username)) {
			if skill["id"] == id && skill["owner"] == identity.Username {
				writeJSON(w, http.StatusOK, map[string]any{"skill": skill})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
	}
}

func agentSkillCreate(usersDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 80<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "技能内容不合法"})
			return
		}
		skill, err := normalizeAgentSkill(input, identity.Username, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		path := agentSkillsPath(usersDir, identity.Username)
		skills := readAgentSkills(path)
		skills = append(skills, skill)
		if err := writeAgentSkills(path, skills); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存技能失败"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"skill": publicAgentSkill(skill)})
	}
}

func agentSkillUpdate(usersDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := req.PathValue("id")
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 80<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "技能内容不合法"})
			return
		}
		path := agentSkillsPath(usersDir, identity.Username)
		skills := readAgentSkills(path)
		for index, current := range skills {
			if current["id"] == id && current["owner"] == identity.Username {
				merged := map[string]any{}
				for key, value := range current {
					merged[key] = value
				}
				for key, value := range input {
					merged[key] = value
				}
				skill, err := normalizeAgentSkill(merged, identity.Username, id)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
					return
				}
				version := intValue(current["version"]) + 1
				skill["version"] = version
				skill["createdAt"] = current["createdAt"]
				skills[index] = skill
				if err := writeAgentSkills(path, skills); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存技能失败"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"skill": publicAgentSkill(skill)})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
	}
}

func agentSkillDelete(usersDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := req.PathValue("id")
		path := agentSkillsPath(usersDir, identity.Username)
		skills := readAgentSkills(path)
		next := make([]map[string]any, 0, len(skills))
		found := false
		for _, skill := range skills {
			if skill["id"] == id && skill["owner"] == identity.Username {
				found = true
				continue
			}
			next = append(next, skill)
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		if err := writeAgentSkills(path, next); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存技能失败"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func agentSkillsPath(usersDir, username string) string {
	return filepath.Join(usersDir, safeAgentUsername(username), "agent-skills.json")
}
func safeAgentUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" || username == "." || username == ".." || strings.ContainsAny(username, `/\\`) {
		return "invalid-user"
	}
	return username
}
func readAgentSkills(path string) []map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	var skills []map[string]any
	if json.Unmarshal(raw, &skills) != nil || skills == nil {
		return []map[string]any{}
	}
	return skills
}
func writeAgentSkills(path string, skills []map[string]any) error {
	raw, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0600)
}
func publicAgentSkill(skill map[string]any) map[string]any {
	output := map[string]any{}
	for _, key := range []string{"id", "source", "name", "description", "category", "inputTemplate", "version", "status"} {
		if value, ok := skill[key]; ok {
			output[key] = value
		}
	}
	return output
}
func normalizeAgentSkill(input map[string]any, owner, id string) (map[string]any, error) {
	name := strings.TrimSpace(toString(input["name"]))
	description := strings.TrimSpace(toString(input["description"]))
	category := strings.TrimSpace(toString(input["category"]))
	template := strings.TrimSpace(toString(input["inputTemplate"]))
	body := strings.TrimSpace(toString(input["body"]))
	if name == "" || len([]rune(name)) > 120 || len([]rune(description)) > 600 || len([]rune(category)) > 80 || len([]rune(template)) > 3000 || body == "" || len([]rune(body)) > 65536 {
		return nil, os.ErrInvalid
	}
	if id == "" {
		id = strings.TrimSpace(toString(input["id"]))
		if id == "" {
			var random [8]byte
			_, _ = rand.Read(random[:])
			id = "skill-" + hex.EncodeToString(random[:])
		}
	}
	if !agentSkillIDPattern.MatchString(id) {
		return nil, os.ErrInvalid
	}
	createdAt := toString(input["createdAt"])
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	return map[string]any{"id": id, "source": "user", "owner": owner, "name": name, "description": description, "category": category, "inputTemplate": template, "body": body, "version": 1, "status": "published", "createdAt": createdAt, "createdBy": owner}, nil
}
func intValue(value any) int {
	switch number := value.(type) {
	case float64:
		return int(number)
	case int:
		return number
	}
	return 0
}
