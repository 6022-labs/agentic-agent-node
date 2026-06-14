package services

import (
	"context"
	"errors"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	vault "github.com/hashicorp/vault/api"
)

type KvV1SecretStorage struct {
	client   *vault.Client
	settings *vault_settings.VaultSettings
}

func NewKvV1SecretStorage(client *vault.Client, settings *vault_settings.VaultSettings) *KvV1SecretStorage {
	return &KvV1SecretStorage{client: client, settings: settings}
}

func (s *KvV1SecretStorage) Write(ctx context.Context, key string, data map[string]string) error {
	payload := make(map[string]any, len(data))
	for k, v := range data {
		payload[k] = v
	}
	return s.client.KVv1(s.settings.KvV1MountPath).Put(ctx, key, payload)
}

func (s *KvV1SecretStorage) Read(ctx context.Context, key string) (map[string]string, error) {
	secret, err := s.client.KVv1(s.settings.KvV1MountPath).Get(ctx, key)
	if err != nil {
		if errors.Is(err, vault.ErrSecretNotFound) {
			return nil, runtime_common_interfaces.ErrSecretNotFound
		}
		var respErr *vault.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return nil, runtime_common_interfaces.ErrSecretNotFound
		}
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, runtime_common_interfaces.ErrSecretNotFound
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
