package external

// New121Provider keeps the 121 gate and endpoint construction explicit at the
// composition boundary while sharing the audited HTTP transport contract.
func New121Provider(endpoint, apiKey string) *HTTPProvider {
	return &HTTPProvider{Provider: Provider121, Endpoint: endpoint, APIKey: apiKey}
}
