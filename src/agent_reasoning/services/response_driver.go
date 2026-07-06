package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ResponseDriver runs the inference loop over a conversation, dedups inputs
// against stored history, and persists new inputs + the final assistant turn
// atomically after the loop succeeds.
type ResponseDriver struct {
	logger                     *zap.Logger
	inferenceLoop              interfaces.InferenceLoopInterface
	mcpConnectionRepository    repositories.MCPConnectionRepositoryInterface
	llmConnectionRepository    repositories.LLMConnectionRepositoryInterface
	agentGuidanceRepository    repositories.AgentGuidanceRepositoryInterface
	mcpHeaderAuthSecretStorage interfaces.MCPHeaderAuthSecretStorageInterface
	conversationHistory        knowledge_interfaces.ConversationHistoryInterface
	agentRepository            identity_repositories.AgentRepositoryInterface
	systemPromptBuilder        interfaces.SystemPromptBuilderInterface
	mcpRequester               interfaces.McpRequesterInterface
	selfToolRegistry           interfaces.SelfToolRegistryInterface
	reasoningTracer            traces_interfaces.ReasoningTracerInterface
}

func NewResponseDriver(
	logger *zap.Logger,
	inferenceLoop interfaces.InferenceLoopInterface,
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface,
	agentGuidanceRepository repositories.AgentGuidanceRepositoryInterface,
	mcpHeaderAuthSecretStorage interfaces.MCPHeaderAuthSecretStorageInterface,
	conversationHistory knowledge_interfaces.ConversationHistoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	systemPromptBuilder interfaces.SystemPromptBuilderInterface,
	mcpRequester interfaces.McpRequesterInterface,
	selfToolRegistry interfaces.SelfToolRegistryInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
) *ResponseDriver {
	return &ResponseDriver{
		logger:                     logger,
		inferenceLoop:              inferenceLoop,
		mcpConnectionRepository:    mcpConnectionRepository,
		llmConnectionRepository:    llmConnectionRepository,
		agentGuidanceRepository:    agentGuidanceRepository,
		mcpHeaderAuthSecretStorage: mcpHeaderAuthSecretStorage,
		conversationHistory:        conversationHistory,
		agentRepository:            agentRepository,
		systemPromptBuilder:        systemPromptBuilder,
		mcpRequester:               mcpRequester,
		selfToolRegistry:           selfToolRegistry,
		reasoningTracer:            reasoningTracer,
	}
}

func (d *ResponseDriver) Drive(
	ctx context.Context,
	conversationId string,
	instructions string,
	inputs []knowledge_interfaces.ConversationMessage,
) (*knowledge_interfaces.ConversationMessage, error) {
	history, err := d.conversationHistory.Read(ctx, conversationId)
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

	messages := d.knowledgeMessagesToChatMessages(combined)

	agent, err := d.loadAgent(ctx)
	if err != nil {
		return nil, err
	}

	llmConnection, err := d.loadLLMConnection(ctx, agent.Id)
	if err != nil {
		return nil, err
	}

	mcpConnections, err := d.loadMCPConnections(ctx, agent.Id)
	if err != nil {
		return nil, err
	}

	mcpTools, err := d.mcpConnectionsToTools(ctx, mcpConnections)
	if err != nil {
		return nil, err
	}

	selfTools := d.selfToolRegistry.ListTools()

	guidance, err := d.loadGuidance(ctx, agent.Id)
	if err != nil {
		return nil, err
	}

	systemPrompt := d.systemPromptBuilder.Build(agent, guidance, selfTools, mcpTools, instructions)
	messages = append([]*interfaces.ChatCompletionMessage{{Role: "system", Content: systemPrompt}}, messages...)

	finalStream, err := d.inferenceLoop.Run(ctx, llmConnection, selfTools, mcpTools, messages)
	if err != nil {
		return nil, fmt.Errorf("inference loop failed: %w", err)
	}

	added := finalStream[len(messages):]
	finalAssistant := d.pickFinalAssistantMessage(added)

	toPersist, err := d.assembleToPersist(newInputs, added)
	if err != nil {
		return nil, err
	}
	if len(toPersist) == 0 {
		return nil, nil
	}

	if _, err := d.conversationHistory.Write(ctx, conversationId, toPersist); err != nil {
		return nil, fmt.Errorf("failed to persist conversation: %w", err)
	}

	return finalAssistant, nil
}

// assembleToPersist stamps strictly increasing timestamps so stored order — and tool-call ↔ result pairing — survives a timestamp-asc reload.
func (d *ResponseDriver) assembleToPersist(
	newInputs []knowledge_interfaces.ConversationMessage,
	added []*interfaces.ChatCompletionMessage,
) ([]knowledge_interfaces.ConversationMessage, error) {
	base := time.Now()
	out := make([]knowledge_interfaces.ConversationMessage, 0, len(newInputs)+len(added))
	for _, m := range newInputs {
		m.Timestamp = base.Add(time.Duration(len(out)) * time.Millisecond)
		out = append(out, m)
	}
	for _, m := range added {
		if m.Content == "" && len(m.ToolCalls) == 0 && m.ToolCallId == "" {
			continue
		}
		km, err := d.chatMessageToKnowledgeMessage(m)
		if err != nil {
			return nil, err
		}
		km.Timestamp = base.Add(time.Duration(len(out)) * time.Millisecond)
		out = append(out, km)
	}
	return out, nil
}

func (d *ResponseDriver) computeHashes(
	conversationId string,
	messages []knowledge_interfaces.ConversationMessage,
) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(messages))
	for _, m := range messages {
		hash, err := utils.ComputeConversationMessageHash(m.Role, m.Content, m.ToolCalls, m.ToolCallId, conversationId)
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

		hash, err := utils.ComputeConversationMessageHash(m.Role, m.Content, m.ToolCalls, m.ToolCallId, conversationId)
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

func (d *ResponseDriver) loadAgent(ctx context.Context) (*identity_models.Agent, error) {
	agent, err := d.agentRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent identity: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent identity is not initialized yet")
	}
	return agent, nil
}

func (d *ResponseDriver) loadLLMConnection(ctx context.Context, agentId uuid.UUID) (*models.LLMConnection, error) {
	llmConnection, err := d.llmConnectionRepository.GetByAgentId(ctx, agentId)
	if err != nil {
		return nil, fmt.Errorf("failed to load llm connection: %w", err)
	}
	if llmConnection == nil {
		return nil, interfaces.ErrLLMConnectionNotConfigured
	}
	return llmConnection, nil
}

func (d *ResponseDriver) loadGuidance(ctx context.Context, agentId uuid.UUID) (string, error) {
	guidance, err := d.agentGuidanceRepository.GetByAgentId(ctx, agentId)
	if err != nil {
		return "", fmt.Errorf("failed to load agent guidance: %w", err)
	}
	if guidance == nil {
		return "", nil
	}
	return guidance.Content, nil
}

func (d *ResponseDriver) mcpConnectionsToTools(
	ctx context.Context,
	mcpConnections []*models.MCPConnection,
) (map[*models.MCPConnection][]*models.Tool, error) {
	mcpTools := make(map[*models.MCPConnection][]*models.Tool, len(mcpConnections))
	for _, conn := range mcpConnections {
		tools, err := d.mcpConnectionToTools(ctx, conn)
		if err != nil {
			return nil, err
		}
		mcpTools[conn] = tools
	}
	return mcpTools, nil
}

func (d *ResponseDriver) mcpConnectionToTools(ctx context.Context, conn *models.MCPConnection) ([]*models.Tool, error) {
	_, span := d.reasoningTracer.StartListTools(ctx, conn.Name)
	defer span.End()

	tools, err := d.mcpRequester.ListTools(conn)
	if err != nil {
		span.Fail(err)
		return nil, fmt.Errorf("failed to list tools for mcp connection %q: %w", conn.Name, err)
	}
	return tools, nil
}

func (d *ResponseDriver) loadMCPConnections(ctx context.Context, agentId uuid.UUID) ([]*models.MCPConnection, error) {
	raw, err := d.mcpConnectionRepository.ListByAgentId(ctx, agentId)
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

func (d *ResponseDriver) knowledgeMessagesToChatMessages(history []knowledge_interfaces.ConversationMessage) []*interfaces.ChatCompletionMessage {
	out := make([]*interfaces.ChatCompletionMessage, 0, len(history))
	for _, m := range history {
		// System rows only feed the session-context overlay — the builder owns the system slot.
		if m.Role == "system" {
			continue
		}
		msg := &interfaces.ChatCompletionMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallId: m.ToolCallId,
		}
		if m.ToolCalls != "" {
			// Tolerate a corrupted or hand-edited row: drop the unreadable tool calls instead of failing the whole conversation.
			if err := json.Unmarshal([]byte(m.ToolCalls), &msg.ToolCalls); err != nil {
				d.logger.Warn("skipping unreadable stored tool calls", zap.String("role", m.Role), zap.Error(err))
				msg.ToolCalls = nil
			}
		}
		out = append(out, msg)
	}
	return out
}

func (d *ResponseDriver) chatMessageToKnowledgeMessage(m *interfaces.ChatCompletionMessage) (knowledge_interfaces.ConversationMessage, error) {
	km := knowledge_interfaces.ConversationMessage{
		Role:       m.Role,
		Content:    m.Content,
		ToolCallId: m.ToolCallId,
	}
	if len(m.ToolCalls) > 0 {
		b, err := json.Marshal(m.ToolCalls)
		if err != nil {
			return knowledge_interfaces.ConversationMessage{}, fmt.Errorf("failed to marshal tool calls: %w", err)
		}
		km.ToolCalls = string(b)
	}
	return km, nil
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
