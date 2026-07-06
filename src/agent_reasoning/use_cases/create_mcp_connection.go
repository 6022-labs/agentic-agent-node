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
	agentRepository identity_repositories.AgentRepositoryInterface,
) *CreateMCPConnection {
	return &CreateMCPConnection{
		AbstractUpsertMCPConnection: NewAbstractUpsertMCPConnection(repository, secretStorage, agentRepository),
	}
}

// Execute creates the row first to get a stable connectionId, then writes
// every header value to Vault under that id. The persisted row only ever
// stores the header *names*; values stay in Vault.
func (u *CreateMCPConnection) Execute(ctx context.Context, request *requests.SaveMCPConnectionRequest) (*responses.MCPConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	keys := models.MCPConnectionHeaderAuthKeys{}
	headerNames := make([]string, 0, len(request.HeaderAuth))
	for k := range request.HeaderAuth {
		headerNames = append(headerNames, k)
	}
	keys.FromArray(headerNames)

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	if err := u.ensureNameAvailable(ctx, agent.Id, request.Name, uuid.Nil); err != nil {
		return nil, err
	}

	created, err := u.repository.Create(ctx, &models.MCPConnection{
		AgentId:        agent.Id,
		Name:           request.Name,
		Url:            request.Url,
		HeaderAuthKeys: keys,
	})
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}

	for k, v := range request.HeaderAuth {
		if err := u.secretStorage.SaveValue(ctx, created.Id, k, v); err != nil {
			return nil, errors.NewInternalError("create_failed", err.Error())
		}
	}
	created.HeaderAuth = request.HeaderAuth
	return responses.ToMCPConnectionResponse(created), nil
}
