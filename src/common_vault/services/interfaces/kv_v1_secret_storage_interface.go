package interfaces

import "context"

// KvV1SecretStorageInterface is the read/write surface over a Vault KV v1
// mount. Returns ErrSecretNotFound (defined in the services package) on Read
// when the key is absent.
type KvV1SecretStorageInterface interface {
	Read(ctx context.Context, key string) (map[string]string, error)
	Write(ctx context.Context, key string, data map[string]string) error
	Delete(ctx context.Context, key string) error
}
