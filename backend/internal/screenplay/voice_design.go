package screenplay

type VoiceDesignDraft struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Prompt      string `json:"prompt"`
	PreviewText string `json:"preview_text"`
	AssetID     string `json:"asset_id,omitempty"`
}
type VoiceDesignState struct {
	Items      []VoiceDesignDraft `json:"items"`
	Model      string             `json:"model"`
	Source     string             `json:"source"`
	RequestKey string             `json:"requestKey"`
	Status     string             `json:"status"`
	Error      string             `json:"error,omitempty"`
}
