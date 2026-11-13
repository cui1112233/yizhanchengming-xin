package screenplay

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// UserStyle belongs only to the authenticated account. Images are protected object references.
type UserStyle struct {
	ID            string    `json:"id"`
	OwnerUsername string    `json:"-"`
	Name          string    `json:"name"`
	Prompt        string    `json:"prompt"`
	Images        []string  `json:"images"`
	Revision      int64     `json:"revision"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreativeStore interface {
	ListUserStyles(context.Context, string) ([]UserStyle, error)
	GetUserStyle(context.Context, string, string) (UserStyle, error)
	SaveUserStyle(context.Context, UserStyle, *int64) (UserStyle, error)
	DeleteUserStyle(context.Context, string, string, int64) error
	SaveCreativeConfig(context.Context, string, string, CreativeConfig, int64) (Project, error)
}

// json-visible metadata never includes system bodies. This envelope is storage-only.
type creativeEnvelope struct {
	Config    CreativeConfig                  `json:"config"`
	Snapshots map[string]storedCreativePrompt `json:"prompt_snapshots"`
}
type storedCreativePrompt struct {
	CreativePromptSnapshot
	Body string `json:"body"`
}

func encodeCreativeConfig(c *CreativeConfig) ([]byte, error) {
	if c == nil {
		return nil, nil
	}
	snapshots := map[string]storedCreativePrompt{}
	for k, v := range c.PromptSnapshots {
		snapshots[k] = storedCreativePrompt{CreativePromptSnapshot: v, Body: v.Body}
	}
	return json.Marshal(creativeEnvelope{Config: *c, Snapshots: snapshots})
}
func decodeCreativeConfig(b []byte) (*CreativeConfig, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var e creativeEnvelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, err
	}
	e.Config.PromptSnapshots = map[string]CreativePromptSnapshot{}
	for k, v := range e.Snapshots {
		p := v.CreativePromptSnapshot
		p.Body = v.Body
		e.Config.PromptSnapshots[k] = p
	}
	return &e.Config, nil
}
func cloneCreativeConfig(c *CreativeConfig) *CreativeConfig {
	b, _ := encodeCreativeConfig(c)
	out, _ := decodeCreativeConfig(b)
	return out
}
func applyCreativeConfig(p *Project, c CreativeConfig) {
	p.CreativeConfig = cloneCreativeConfig(&c)
	if c.Style == nil {
		p.ArtStyle = ""
		p.SelectedStyle = ""
		p.StyleDesc = ""
	} else {
		p.ArtStyle = c.Style.Name
		p.SelectedStyle = c.Style.ID
		p.StyleDesc = c.Style.Prompt
	}
	if strings.TrimSpace(c.ScriptPromptID) != "" {
		p.PromptID = c.ScriptPromptID
	}
}
