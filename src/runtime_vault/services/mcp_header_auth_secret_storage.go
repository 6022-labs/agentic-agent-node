package services

import (
	"context"
	"errors"
	"fmt"

	vault_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services/interfaces"
	"github.com/google/uuid"
)

const (
	mcpHeaderAuthSecretKeyTemplate = "agent/mcp_auth_header_secret/%s/%s"
	mcpHeaderAuthFieldValue        = "value"
)

type MCPHeaderAuthSecretStorage struct {
	storage vault_interfaces.KvV1SecretStorageInterface
}

func NewMCPHeaderAuthSecretStorage(storage vault_interfaces.KvV1SecretStorageInterface) *MCPHeaderAuthSecretStorage {
	return &MCPHeaderAuthSecretStorage{storage: storage}
}

func (s *MCPHeaderAuthSecretStorage) GetValue(ctx context.Context, connectionId uuid.UUID, key string) (string, bool, error) {
	data, err := s.storage.Read(ctx, secretPath(connectionId, key))
	if errors.Is(err, ErrSecretNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	value, ok := data[mcpHeaderAuthFieldValue]
	if !ok {
		return "", false, fmt.Errorf("mcp header auth secret missing %q field", mcpHeaderAuthFieldValue)
	}
	return value, true, nil
}

func (s *MCPHeaderAuthSecretStorage) SaveValue(ctx context.Context, connectionId uuid.UUID, key string, value string) error {
	if len(key) == 0 {
		return errors.New("header key cannot be empty")
	}
	if len(value) == 0 {
		return errors.New("header value cannot be empty")
	}
	return s.storage.Write(ctx, secretPath(connectionId, key), map[string]string{mcpHeaderAuthFieldValue: value})
}

func (s *MCPHeaderAuthSecretStorage) DeleteValue(ctx context.Context, connectionId uuid.UUID, key string) error {
	return s.storage.Delete(ctx, secretPath(connectionId, key))
}

func secretPath(connectionId uuid.UUID, key string) string {
	return fmt.Sprintf(mcpHeaderAuthSecretKeyTemplate, connectionId, key)
}
