package novelfetchworkshop

// ProcessedBody contains the complete upstream text and the bounded text used by processing.
type ProcessedBody struct {
	OriginalRaw      string `json:"originalRaw"`
	ProcessedText    string `json:"processedText"`
	OriginalRawChars int    `json:"originalRawChars"`
	ProcessedChars   int    `json:"processedChars"`
	MaxTxt           int    `json:"maxTxt"`
}

func ProcessRawText(originalRaw string, maxTxt int) ProcessedBody {
	if maxTxt < 0 {
		maxTxt = 0
	}

	processed := []rune(originalRaw)
	if len(processed) > maxTxt {
		processed = processed[:maxTxt]
	}

	return ProcessedBody{
		OriginalRaw:      originalRaw,
		ProcessedText:    string(processed),
		OriginalRawChars: len([]rune(originalRaw)),
		ProcessedChars:   len(processed),
		MaxTxt:           maxTxt,
	}
}
