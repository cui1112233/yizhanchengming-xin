package httpapi

import (
	"errors"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
)

type screenplayStyle struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PreviewURL string `json:"previewUrl"`
}

func screenplayStyles(catalog *prompt.Catalog) ([]screenplayStyle, error) {
	styles := []screenplayStyle{}
	if catalog == nil {
		return styles, nil
	}
	presets, err := catalog.Published("batch-factory")
	if err != nil {
		return nil, err
	}
	for _, preset := range presets {
		if preset.Slot != nil && *preset.Slot == "batch.comic-style" {
			styles = append(styles, screenplayStyle{ID: preset.ID, Name: preset.Name, PreviewURL: preset.PreviewURL})
		}
	}
	return styles, nil
}

// Freeze the selected preset body in the project. Browser-supplied descriptions
// are never trusted. Old name-only projects retain their existing style.
func resolveScreenplayStyle(catalog *prompt.Catalog, selected, name string, current *screenplay.Project) (screenplay.Project, error) {
	selected, name = strings.TrimSpace(selected), strings.TrimSpace(name)
	if selected == "" {
		if current != nil && (name == "" || name == current.ArtStyle) {
			return *current, nil
		}
		return screenplay.Project{ArtStyle: name}, nil
	}
	if catalog != nil {
		preset, found, err := catalog.PublishedBody("batch-factory", selected)
		if err != nil {
			return screenplay.Project{}, err
		}
		if found && preset.Slot == "batch.comic-style" && strings.TrimSpace(preset.Body) != "" {
			return screenplay.Project{ArtStyle: preset.Name, SelectedStyle: preset.ID, StyleDesc: preset.Body}, nil
		}
	}
	if current != nil && selected == current.ArtStyle && (current.SelectedStyle == "" || current.SelectedStyle == current.ArtStyle) && (name == "" || name == current.ArtStyle) {
		return *current, nil
	}
	return screenplay.Project{}, errors.New("所选画风已失效或不可用于漫剧，请重新选择画风")
}
