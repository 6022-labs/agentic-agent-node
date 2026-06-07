package services

import (
	"context"
	"errors"
	"fmt"

	vault_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services/interfaces"
)

const (
	llmSecretKey   = "agent/llm_secret"
	llmFieldApiKey = "apiKey"
)

type LLMSecretStorage struct {
	storage vault_interfaces.KvV1SecretStorageInterface
}

func NewLLMSecretStorage(storage vault_interfaces.KvV1SecretStorageInterface) *LLMSecretStorage {
	return &LLMSecretStorage{storage: storage}
}

func (s *LLMSecretStorage) Get(ctx context.Context) (string, bool, error) {
	data, err := s.storage.Read(ctx, llmSecretKey)
	if errors.Is(err, ErrSecretNotFound) {
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
	return s.storage.Write(ctx, llmSecretKey, map[string]string{llmFieldApiKey: apiKey})
}

func (s *LLMSecretStorage) Delete(ctx context.Context) error {
	return s.storage.Delete(ctx, llmSecretKey)
}
