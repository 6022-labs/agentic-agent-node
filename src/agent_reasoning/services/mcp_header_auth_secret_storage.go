package services

import (
	"context"
	"errors"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

const (
	mcpHeaderAuthSecretKeyTemplate = "mcp_auth_header_secret/%s/%s"
	mcpHeaderAuthFieldValue        = "value"
)

type MCPHeaderAuthSecretStorage struct {
	storage               runtime_common_interfaces.SecretStorageInterface
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface
}

func NewMCPHeaderAuthSecretStorage(
	storage runtime_common_interfaces.SecretStorageInterface,
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface,
) *MCPHeaderAuthSecretStorage {
	return &MCPHeaderAuthSecretStorage{storage: storage, agentScopedKeyBuilder: agentScopedKeyBuilder}
}

func (s *MCPHeaderAuthSecretStorage) GetValue(
	ctx context.Context,
	connectionId uuid.UUID,
	key string,
) (string, bool, error) {
	secretKey, err := s.secretPath(ctx, connectionId, key)
	if err != nil {
		return "", false, err
	}
	data, err := s.storage.Read(ctx, secretKey)
	if errors.Is(err, apperrors.ErrSecretNotFound) {
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

func (s *MCPHeaderAuthSecretStorage) SaveValue(
	ctx context.Context,
	connectionId uuid.UUID,
	key string,
	value string,
) error {
	if len(key) == 0 {
		return errors.New("header key cannot be empty")
	}
	if len(value) == 0 {
		return errors.New("header value cannot be empty")
	}
	secretKey, err := s.secretPath(ctx, connectionId, key)
	if err != nil {
		return err
	}
	return s.storage.Write(ctx, secretKey, map[string]string{mcpHeaderAuthFieldValue: value})
}

func (s *MCPHeaderAuthSecretStorage) DeleteValue(ctx context.Context, connectionId uuid.UUID, key string) error {
	secretKey, err := s.secretPath(ctx, connectionId, key)
	if err != nil {
		return err
	}
	return s.storage.Delete(ctx, secretKey)
}

func (s *MCPHeaderAuthSecretStorage) secretPath(
	ctx context.Context,
	connectionId uuid.UUID,
	key string,
) (string, error) {
	return s.agentScopedKeyBuilder.ScopedKey(ctx, fmt.Sprintf(mcpHeaderAuthSecretKeyTemplate, connectionId, key))
}
