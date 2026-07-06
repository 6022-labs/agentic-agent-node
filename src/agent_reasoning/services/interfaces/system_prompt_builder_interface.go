package interfaces

import (
	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type SystemPromptBuilderInterface interface {
	// Build composes the layered prompt (identity → soul → behavior → guidance →
	// tools → session overlay → timestamp); rebuilt per request, never persisted.
	Build(
		agent *identity_models.Agent,
		guidance string,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		sessionContext string,
	) string
}
