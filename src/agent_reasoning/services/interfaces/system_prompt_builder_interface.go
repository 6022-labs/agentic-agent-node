package interfaces

import (
	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type SystemPromptBuilderInterface interface {
	// Build composes the layered prompt (identity → soul → behavior → guidance →
	// tools → session overlay → timestamp); rebuilt per request, never persisted.
	Build(
		agent *identity_models.Agent,
		addresses map[identity_enums.AddressType]string,
		guidance string,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		sessionContext string,
	) string
}
