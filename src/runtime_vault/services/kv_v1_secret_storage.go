package services

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_vault/metrics/interfaces"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	vault "github.com/hashicorp/vault/api"
)

type KvV1SecretStorage struct {
	client   *vault.Client
	settings *vault_settings.VaultSettings
	metrics  metrics_interfaces.VaultMetricsInterface
}

func NewKvV1SecretStorage(
	client *vault.Client,
	settings *vault_settings.VaultSettings,
	metrics metrics_interfaces.VaultMetricsInterface,
) *KvV1SecretStorage {
	return &KvV1SecretStorage{client: client, settings: settings, metrics: metrics}
}

func (s *KvV1SecretStorage) recordOperation(ctx context.Context, operation string, start time.Time, err *error) {
	// ErrSecretNotFound is an expected "miss", not an operational failure.
	var operationErr error
	if err != nil && *err != nil && !errors.Is(*err, apperrors.ErrSecretNotFound) {
		operationErr = *err
	}
	s.metrics.RecordOperation(ctx, operation, s.settings.KvV1MountPath, time.Since(start), operationErr)
}

func (s *KvV1SecretStorage) Write(ctx context.Context, key string, data map[string]string) (err error) {
	defer s.recordOperation(ctx, "write", time.Now(), &err)

	payload := make(map[string]any, len(data))
	for k, v := range data {
		payload[k] = v
	}
	return s.client.KVv1(s.settings.KvV1MountPath).Put(ctx, key, payload)
}

func (s *KvV1SecretStorage) Read(ctx context.Context, key string) (result map[string]string, err error) {
	defer s.recordOperation(ctx, "read", time.Now(), &err)

	secret, err := s.client.KVv1(s.settings.KvV1MountPath).Get(ctx, key)
	if err != nil {
		if errors.Is(err, vault.ErrSecretNotFound) {
			return nil, apperrors.ErrSecretNotFound
		}
		var respErr *vault.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == 404 {
			return nil, apperrors.ErrSecretNotFound
		}
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, apperrors.ErrSecretNotFound
	}

	out := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		if str, ok := v.(string); ok {
			out[k] = str
		}
	}
	return out, nil
}

func (s *KvV1SecretStorage) Delete(ctx context.Context, key string) (err error) {
	defer s.recordOperation(ctx, "delete", time.Now(), &err)

	err = s.client.KVv1(s.settings.KvV1MountPath).Delete(ctx, key)
	return err
}
