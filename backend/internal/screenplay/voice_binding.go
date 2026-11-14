package screenplay

type VoiceBinding struct {
	MediaItemID      string `json:"media_item_id"`
	VoiceID          string `json:"voice_id"`
	Engine           string `json:"engine"`
	ReferenceOnly    bool   `json:"reference_only"`
	Source           string `json:"source"`
	LibraryVoiceName string `json:"library_voice_name,omitempty"`
}
