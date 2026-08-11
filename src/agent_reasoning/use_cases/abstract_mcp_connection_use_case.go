package use_cases

import (
	"context"
	stderrors "errors"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// credentialsRejectedMessage reaches the console verbatim, so it names what the user can fix.
const credentialsRejectedMessage = "the server rejected these credentials — check the API key"

// AbstractMCPConnectionUseCase carries the persistence, secret storage and
// requester every MCP connection use case needs, plus the shared server probe.
type AbstractMCPConnectionUseCase struct {
	*AbstractReasoningUseCase
	mcpConnectionRepository    repositories.MCPConnectionRepositoryInterface
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
	mcpRequester               reasoning_interfaces.McpRequesterInterface
}

func NewAbstractMCPConnectionUseCase(
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	mcpRequester reasoning_interfaces.McpRequesterInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *AbstractMCPConnectionUseCase {
	return &AbstractMCPConnectionUseCase{
		AbstractReasoningUseCase:   NewAbstractReasoningUseCase(agentRepository),
		mcpConnectionRepository:    mcpConnectionRepository,
		mcpHeaderAuthSecretStorage: mcpHeaderAuthSecretStorage,
		mcpRequester:               mcpRequester,
	}
}

// listTools probes the server, blaming the headers when it refused them and the url otherwise.
func (u *AbstractMCPConnectionUseCase) listTools(
	ctx context.Context,
	mcpConnection *models.MCPConnection,
) ([]*models.Tool, error) {
	tools, err := u.mcpRequester.ListTools(ctx, mcpConnection)
	if err != nil {
		if stderrors.Is(err, reasoning_interfaces.ErrMCPUnauthorized) {
			return nil, errors.NewValidationError("headerAuth", credentialsRejectedMessage)
		}
		return nil, errors.NewValidationError("url", "failed to list tools: "+err.Error())
	}
	return tools, nil
}
