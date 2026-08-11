package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

type CreateMCPConnection struct {
	*AbstractUpsertMCPConnection
}

func NewCreateMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	mcpRequester reasoning_interfaces.McpRequesterInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *CreateMCPConnection {
	return &CreateMCPConnection{
		AbstractUpsertMCPConnection: NewAbstractUpsertMCPConnection(
			repository,
			secretStorage,
			mcpRequester,
			agentRepository,
		),
	}
}

// Execute creates the row first to get a stable connectionId, then writes
// every header value to Vault under that id. The persisted row only ever
// stores the header *names*; values stay in Vault.
func (u *CreateMCPConnection) Execute(
	ctx context.Context,
	request *requests.CreateMCPConnectionRequest,
) (*responses.MCPConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	if err := u.ensureNameAvailable(ctx, agent.Id, request.Name, uuid.Nil); err != nil {
		return nil, err
	}

	if err := u.probeListTools(ctx, request.Name, request.Url, request.HeaderAuth); err != nil {
		return nil, err
	}

	created, err := u.mcpConnectionRepository.Create(ctx, &models.MCPConnection{
		AgentId:             agent.Id,
		Name:                request.Name,
		Url:                 request.Url,
		AllowedToolPatterns: request.AllowedToolPatterns,
		HeaderAuthKeys:      u.headerAuthKeys(request),
	})
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}

	if err := u.saveHeaderValues(ctx, created.Id, request.HeaderAuth); err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	created.HeaderAuth = request.HeaderAuth
	return responses.ToMCPConnectionResponse(created), nil
}
