package unifiedsettings

import "time"

type Settings struct {
	Production              map[string]any `json:"production,omitempty"`
	Publishing              map[string]any `json:"publishing,omitempty"`
	Website121              map[string]any `json:"website121,omitempty"`
	StyleTypes              map[string]any `json:"styleTypes,omitempty"`
	ProcessingRulePromptRef string         `json:"processingRulePromptRef,omitempty"`
	KnowledgePromptRef      string         `json:"knowledgePromptRef,omitempty"`
}

type VersionProfile struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"projectId"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Settings  Settings  `json:"settings"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type Current struct {
	ProjectID int64          `json:"projectId"`
	Effective Settings       `json:"effective"`
	Project   Settings       `json:"project"`
	Profile   VersionProfile `json:"profile"`
	Priority  []string       `json:"priority"`
}
