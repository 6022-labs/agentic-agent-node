package services

import (
	"context"
	"fmt"
	"time"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"go.uber.org/zap"
)

// ResponseDriver runs the inference loop over a conversation, dedups inputs
// against stored history, and persists new inputs + the final assistant turn
// atomically after the loop succeeds.
type ResponseDriver struct {
	logger                     *zap.Logger
	inferenceLoop              interfaces.InferenceLoopInterface
	mcpConnectionRepository    repositories.MCPConnectionRepositoryInterface
	mcpHeaderAuthSecretStorage interfaces.MCPHeaderAuthSecretStorageInterface
	conversationHistory        knowledge_interfaces.ConversationHistoryInterface
}

func NewResponseDriver(
	logger *zap.Logger,
	inferenceLoop interfaces.InferenceLoopInterface,
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	mcpHeaderAuthSecretStorage interfaces.MCPHeaderAuthSecretStorageInterface,
	conversationHistory knowledge_interfaces.ConversationHistoryInterface,
) *ResponseDriver {
	return &ResponseDriver{
		logger:                     logger,
		inferenceLoop:              inferenceLoop,
		mcpConnectionRepository:    mcpConnectionRepository,
		mcpHeaderAuthSecretStorage: mcpHeaderAuthSecretStorage,
		conversationHistory:        conversationHistory,
	}
}

func (d *ResponseDriver) Drive(
	ctx context.Context,
	conversationId string,
	inputs []knowledge_interfaces.ConversationMessage,
) (*knowledge_interfaces.ConversationMessage, error) {
	history, err := d.conversationHistory.Read(conversationId)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation history: %w", err)
	}

	historyHashes, err := d.computeHashes(conversationId, history)
	if err != nil {
		return nil, err
	}

	newInputs, err := d.dedupInputsAgainstHistory(conversationId, inputs, historyHashes)
	if err != nil {
		return nil, err
	}

	combined := make([]knowledge_interfaces.ConversationMessage, 0, len(history)+len(newInputs))
	combined = append(combined, history...)
	combined = append(combined, newInputs...)

	messages := make([]*interfaces.ChatCompletionMessage, 0, len(combined))
	for _, m := range combined {
		messages = append(messages, &interfaces.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	mcpConnections, err := d.loadMCPConnections(ctx)
	if err != nil {
		return nil, err
	}

	finalStream, err := d.inferenceLoop.Run(ctx, mcpConnections, messages)
	if err != nil {
		return nil, fmt.Errorf("inference loop failed: %w", err)
	}

	finalAssistant := d.pickFinalAssistantMessage(finalStream[len(messages):])

	toPersist := make([]knowledge_interfaces.ConversationMessage, 0, len(newInputs)+1)
	toPersist = append(toPersist, newInputs...)
	if finalAssistant != nil {
		toPersist = append(toPersist, *finalAssistant)
	}

	if len(toPersist) == 0 {
		return nil, nil
	}

	persisted, err := d.conversationHistory.Write(conversationId, toPersist)
	if err != nil {
		return nil, fmt.Errorf("failed to persist conversation: %w", err)
	}

	if finalAssistant == nil {
		return nil, nil
	}
	if len(persisted) > 0 {
		last := persisted[len(persisted)-1]
		return &last, nil
	}
	return finalAssistant, nil
}

func (d *ResponseDriver) computeHashes(
	conversationId string,
	messages []knowledge_interfaces.ConversationMessage,
) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(messages))
	for _, m := range messages {
		hash, err := utils.ComputeConversationMessageHash(m.Role, m.Author, m.Platform, m.Content, conversationId)
		if err != nil {
			return nil, fmt.Errorf("failed to hash history message: %w", err)
		}
		out[hash] = struct{}{}
	}
	return out, nil
}

// dedupInputsAgainstHistory demojizes each input then drops ones whose hash
// already lives in `historyHashes`. Hash is post-demojize so it matches what
// knowledge persists.
func (d *ResponseDriver) dedupInputsAgainstHistory(
	conversationId string,
	inputs []knowledge_interfaces.ConversationMessage,
	historyHashes map[string]struct{},
) ([]knowledge_interfaces.ConversationMessage, error) {
	out := make([]knowledge_interfaces.ConversationMessage, 0, len(inputs))
	for _, m := range inputs {
		m.Content = utils.DemojizeContent(m.Content)

		hash, err := utils.ComputeConversationMessageHash(m.Role, m.Author, m.Platform, m.Content, conversationId)
		if err != nil {
			return nil, fmt.Errorf("failed to hash input message: %w", err)
		}
		if _, hit := historyHashes[hash]; hit {
			continue
		}
		// Also dedupe between inputs themselves.
		historyHashes[hash] = struct{}{}
		out = append(out, m)
	}
	return out, nil
}

func (d *ResponseDriver) loadMCPConnections(ctx context.Context) ([]*models.MCPConnection, error) {
	raw, err := d.mcpConnectionRepository.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list mcp connections: %w", err)
	}

	out := make([]*models.MCPConnection, 0, len(raw))
	for i := range raw {
		c := raw[i]
		c.HeaderAuth = make(map[string]string, len(c.HeaderAuthKeys))
		for _, key := range c.HeaderAuthKeys {
			value, found, err := d.mcpHeaderAuthSecretStorage.GetValue(ctx, c.Id, string(key))
			if err != nil {
				return nil, fmt.Errorf("failed to load header auth %q for mcp connection %q: %w", key, c.Name, err)
			}
			if found {
				c.HeaderAuth[string(key)] = value
			}
		}
		out = append(out, &c)
	}
	return out, nil
}

func (d *ResponseDriver) pickFinalAssistantMessage(added []*interfaces.ChatCompletionMessage) *knowledge_interfaces.ConversationMessage {
	for i := len(added) - 1; i >= 0; i-- {
		m := added[i]
		if m.Role == "assistant" && m.Content != "" {
			return &knowledge_interfaces.ConversationMessage{
				Role:      m.Role,
				Content:   m.Content,
				Timestamp: time.Now(),
			}
		}
	}
	return nil
}
