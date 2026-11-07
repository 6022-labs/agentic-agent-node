package interfaces

import (
	"context"
)

// SecretStorageInterface is the backend-agnostic read/write surface for secrets;
// KvV1SecretStorage (Vault) is one implementation.
type SecretStorageInterface interface {
	// Read returns the secret as a flat string map, or ErrSecretNotFound if absent.
	Read(ctx context.Context, key string) (map[string]string, error)
	Write(ctx context.Context, key string, data map[string]string) error
	Delete(ctx context.Context, key string) error
}
