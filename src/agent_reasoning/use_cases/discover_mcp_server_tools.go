package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/samber/lo"
)

// candidateConnectionName labels a not-yet-saved server in the metrics the requester records.
const candidateConnectionName = "candidate"

type DiscoverMCPServerTools struct {
	*AbstractMCPConnectionUseCase
}

func NewDiscoverMCPServerTools(
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	mcpRequester reasoning_interfaces.McpRequesterInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *DiscoverMCPServerTools {
	return &DiscoverMCPServerTools{
		AbstractMCPConnectionUseCase: NewAbstractMCPConnectionUseCase(
			mcpConnectionRepository,
			mcpHeaderAuthSecretStorage,
			mcpRequester,
			agentResolver,
		),
	}
}

// Execute lists what a server exposes, so the console can offer its tools for selection.
func (u *DiscoverMCPServerTools) Execute(
	ctx context.Context,
	request *requests.DiscoverMCPServerToolsRequest,
) (*responses.MCPServerToolsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	target, err := u.resolveTarget(ctx, request)
	if err != nil {
		return nil, err
	}

	tools, err := u.listTools(ctx, target)
	if err != nil {
		return nil, err
	}

	return responses.ToMCPServerToolsResponse(tools), nil
}

// resolveTarget builds the connection to probe: the request wins, and whatever it
// omits comes from the stored connection — its url, and its header values from Vault.
func (u *DiscoverMCPServerTools) resolveTarget(
	ctx context.Context,
	request *requests.DiscoverMCPServerToolsRequest,
) (*models.MCPConnection, error) {
	if request.ConnectionId == nil {
		return &models.MCPConnection{
			Name:       candidateConnectionName,
			Url:        lo.FromPtr(request.Url),
			HeaderAuth: request.HeaderAuth,
		}, nil
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	stored, err := u.mcpConnectionRepository.GetById(ctx, agent.Id, *request.ConnectionId)
	if err != nil {
		return nil, errors.NewInternalError("list_mcp_server_tools_failed", err)
	}
	if stored == nil {
		return nil, errors.NewNotFoundError("mcp_connection_not_found", "")
	}

	headers, err := u.resolveHeaderValues(ctx, stored, request.HeaderAuth)
	if err != nil {
		return nil, err
	}

	return &models.MCPConnection{
		Name:       stored.Name,
		Url:        lo.FromPtrOr(request.Url, stored.Url),
		HeaderAuth: headers,
	}, nil
}

func (u *DiscoverMCPServerTools) resolveHeaderValues(
	ctx context.Context,
	stored *models.MCPConnection,
	requested map[string]string,
) (map[string]string, error) {
	resolved := make(map[string]string, len(stored.HeaderAuthKeys)+len(requested))
	for _, key := range stored.HeaderAuthKeys {
		value, found, err := u.mcpHeaderAuthSecretStorage.GetValue(ctx, stored.Id, string(key))
		if err != nil {
			return nil, errors.NewInternalError("list_mcp_server_tools_failed", err)
		}
		if found {
			resolved[string(key)] = value
		}
	}
	for key, value := range requested {
		if value != "" {
			resolved[key] = value
		}
	}
	return resolved, nil
}
