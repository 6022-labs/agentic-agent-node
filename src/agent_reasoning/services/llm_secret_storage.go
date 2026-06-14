package services

import (
	"context"
	"errors"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
)

const (
	llmSecretKey   = "llm_secret"
	llmFieldApiKey = "apiKey"
)

type LLMSecretStorage struct {
	storage               runtime_common_interfaces.SecretStorageInterface
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface
}

func NewLLMSecretStorage(
	storage runtime_common_interfaces.SecretStorageInterface,
	agentScopedKeyBuilder identity_interfaces.AgentScopedKeyBuilderInterface,
) *LLMSecretStorage {
	return &LLMSecretStorage{storage: storage, agentScopedKeyBuilder: agentScopedKeyBuilder}
}

func (s *LLMSecretStorage) Get(ctx context.Context) (string, bool, error) {
	key, err := s.agentScopedKeyBuilder.ScopedKey(ctx, llmSecretKey)
	if err != nil {
		return "", false, err
	}
	data, err := s.storage.Read(ctx, key)
	if errors.Is(err, runtime_common_interfaces.ErrSecretNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	apiKey, ok := data[llmFieldApiKey]
	if !ok || len(apiKey) == 0 {
		return "", false, fmt.Errorf("llm secret missing %q field", llmFieldApiKey)
	}
	return apiKey, true, nil
}

func (s *LLMSecretStorage) Save(ctx context.Context, apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("apiKey cannot be empty")
	}
	key, err := s.agentScopedKeyBuilder.ScopedKey(ctx, llmSecretKey)
	if err != nil {
		return err
	}
	return s.storage.Write(ctx, key, map[string]string{llmFieldApiKey: apiKey})
}

func (s *LLMSecretStorage) Delete(ctx context.Context) error {
	key, err := s.agentScopedKeyBuilder.ScopedKey(ctx, llmSecretKey)
	if err != nil {
		return err
	}
	return s.storage.Delete(ctx, key)
}
