package agentchat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill is the Tier-1 + Tier-2 record. Only id and description are shown to
// the model up front; the body is fetched through skill_view.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Body        string `json:"body"`
	// Command is optional. run_skill refuses to execute a skill that has not
	// declared an explicit command, so a data-only skill can never spawn a
	// process by accident.
	Command string `json:"command"`
	Source  string `json:"source"`
	Owner   string `json:"owner"`
	Status  string `json:"status"`
}

// SkillCatalog is an immutable snapshot resolved per request.
type SkillCatalog struct {
	items []Skill
	byID  map[string]Skill
}

// NewSkillCatalog builds a catalog, keeping the first occurrence of each id so
// system skills win over user skills (the original "同 id 时内置优先" rule).
func NewSkillCatalog(items []Skill) *SkillCatalog {
	catalog := &SkillCatalog{items: []Skill{}, byID: map[string]Skill{}}
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, exists := catalog.byID[id]; exists {
			continue
		}
		item.ID = id
		catalog.byID[id] = item
		catalog.items = append(catalog.items, item)
	}
	sort.SliceStable(catalog.items, func(i, j int) bool { return catalog.items[i].ID < catalog.items[j].ID })
	return catalog
}

// List returns the catalog in stable id order.
func (c *SkillCatalog) List() []Skill {
	if c == nil {
		return []Skill{}
	}
	return append([]Skill(nil), c.items...)
}

// Get resolves one skill by id.
func (c *SkillCatalog) Get(id string) (Skill, bool) {
	if c == nil {
		return Skill{}, false
	}
	skill, ok := c.byID[strings.TrimSpace(id)]
	return skill, ok
}

// LoadSkillFiles reads the JSON arrays used by the existing Agent skill store
// (system agent-skills.json and <user>/agent-skills.json). Unreadable or
// malformed files are skipped, matching the original "跳过损坏 Skill" behaviour.
func LoadSkillFiles(paths ...string) []Skill {
	skills := []Skill{}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var records []map[string]any
		if json.Unmarshal(raw, &records) != nil {
			continue
		}
		for _, record := range records {
			skill := Skill{
				ID:          textValue(record["id"]),
				Name:        textValue(record["name"]),
				Description: textValue(record["description"]),
				Category:    textValue(record["category"]),
				Body:        textValue(record["body"]),
				Command:     textValue(record["command"]),
				Source:      textValue(record["source"]),
				Owner:       textValue(record["owner"]),
				Status:      textValue(record["status"]),
			}
			if skill.ID == "" || skill.Body == "" {
				continue
			}
			if skill.Status != "" && skill.Status != "published" {
				continue
			}
			skills = append(skills, skill)
		}
		// Also pick up packaged skills: <dir>/installed/<id>/SKILL.md
		root := filepath.Dir(path)
		skills = append(skills, loadPackagedSkills(filepath.Join(root, "installed"))...)
	}
	return skills
}

func loadPackagedSkills(installedDir string) []Skill {
	entries, err := os.ReadDir(installedDir)
	if err != nil {
		return nil
	}
	skills := make([]Skill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		body, err := os.ReadFile(filepath.Join(installedDir, entry.Name(), "SKILL.md"))
		if err != nil {
			continue
		}
		skills = append(skills, Skill{
			ID:     entry.Name(),
			Name:   entry.Name(),
			Body:   string(body),
			Source: "installed",
			Status: "published",
		})
	}
	return skills
}

func textValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	}
	return ""
}
