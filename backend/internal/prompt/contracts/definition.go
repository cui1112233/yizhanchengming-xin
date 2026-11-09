package contracts

// Definition is the backend-owned identity and behavior contract for one
// system prompt. Callers persist ID and Version instead of copying prompt text
// into frontend state or task records.
type Definition struct {
	ID             string
	Version        int
	SystemPrompt   string
	RequiredInputs []string
	OutputContract string
	Errors         map[string]string
}
