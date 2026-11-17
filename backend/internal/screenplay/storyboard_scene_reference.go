package screenplay

import "strings"

// Only a bound chapter, or a chapter whose script matches this input, may
// supply an independent novel reference. Direct input remains authoritative.
func storyboardReferenceOriginal(req StoryboardRequest, state *AuthoringState, episode episodeInput) string {
	if state == nil || req.UseOriginalContentDirectly {
		return episode.Content
	}
	for i, c := range state.Chapters {
		if (req.ChapterID != "" && c.ID != req.ChapterID) || (req.ChapterID == "" && i+1 != episode.No) {
			continue
		}
		if strings.TrimSpace(c.SourceText) != "" && (req.ChapterID != "" || strings.TrimSpace(c.ScriptText) == strings.TrimSpace(episode.Content)) {
			return c.SourceText
		}
		break
	}
	return episode.Content
}
