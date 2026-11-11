package external

// NewYadiProvider is the independently gated Yadi composition entry point.
func NewYadiProvider(endpoint, apiKey string) *HTTPProvider {
	return &HTTPProvider{Provider: ProviderYadi, Endpoint: endpoint, APIKey: apiKey}
}
