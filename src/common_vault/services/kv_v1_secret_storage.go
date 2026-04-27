package services

import (
	"context"
	"errors"

	vault_settings "github.com/6022-labs/agentic-agent-node/src/common_vault/settings"
	vault "github.com/hashicorp/vault/api"
)

var ErrSecretNotFound = errors.New("secret not found")

type KvV1SecretStorage struct {
	client   *vault.Client
	settings *vault_settings.VaultSettings
}

func NewKvV1SecretStorage(client *vault.Client, settings *vault_settings.VaultSettings) *KvV1SecretStorage {
	return &KvV1SecretStorage{client: client, settings: settings}
}

// Write upserts a KV v1 secret as a flat string map.
func (s *KvV1SecretStorage) Write(ctx context.Context, key string, data map[string]string) error {
	payload := make(map[string]any, len(data))
	for k, v := range data {
		payload[k] = v
	}
	return s.client.KVv1(s.settings.KvV1MountPath).Put(ctx, key, payload)
}

// Read returns the secret as a flat string map. Returns ErrSecretNotFound if absent.
func (s *KvV1SecretStorage) Read(ctx context.Context, key string) (map[string]string, error) {
	secret, err := s.client.KVv1(s.settings.KvV1MountPath).Get(ctx, key)
	if err != nil {
		if errors.Is(err, vault.ErrSecretNotFound) {
			return nil, ErrSecretNotFound
		}
		var respErr *vault.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return nil, ErrSecretNotFound
		}
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, ErrSecretNotFound
	}

	out := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		if str, ok := v.(string); ok {
			out[k] = str
		}
	}
	return out, nil
}

func (s *KvV1SecretStorage) Delete(ctx context.Context, key string) error {
	return s.client.KVv1(s.settings.KvV1MountPath).Delete(ctx, key)
}
