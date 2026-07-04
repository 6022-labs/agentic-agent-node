package interfaces

import (
	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

type SystemPromptBuilderInterface interface {
	// Build composes the layered prompt (identity → soul → behavior → guidance →
	// tools → session overlay → timestamp); rebuilt per request, never persisted.
	Build(
		agent *identity_models.Agent,
		guidance string,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		sessionContext string,
	) string
}
