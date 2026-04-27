package interfaces

import "context"

// LLMSecretStorageInterface persists the LLM provider's api key. The key is
// always stored at a fixed Vault path (one agent per container).
type LLMSecretStorageInterface interface {
	// Get returns the api key, or ("", false, nil) if none has been saved.
	Get(ctx context.Context) (string, bool, error)
	Save(ctx context.Context, apiKey string) error
	// Delete clears the api key. No-op if none is set.
	Delete(ctx context.Context) error
}
