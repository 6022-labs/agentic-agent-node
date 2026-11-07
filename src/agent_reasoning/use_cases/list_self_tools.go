package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListSelfTools struct {
	*AbstractReasoningUseCase
	selfToolRepository repositories.SelfToolRepositoryInterface
	selfToolRegistry   interfaces.SelfToolRegistryInterface
}

func NewListSelfTools(
	selfToolRepository repositories.SelfToolRepositoryInterface,
	selfToolRegistry interfaces.SelfToolRegistryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ListSelfTools {
	return &ListSelfTools{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		selfToolRepository:       selfToolRepository,
		selfToolRegistry:         selfToolRegistry,
	}
}

// Execute lists the self tool rows with their descriptions; rows are reconciled at startup.
func (u *ListSelfTools) Execute(ctx context.Context) (*responses.SelfToolListResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := u.selfToolRepository.ListByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}

	descriptionByName := make(map[string]string)
	for _, definition := range u.selfToolRegistry.ListTools() {
		descriptionByName[definition.Name] = definition.Description
	}
	groupByName := u.selfToolRegistry.ListGroupNames()

	// Post-reconcile every row matches a declared tool.
	tools := make([]responses.SelfToolResponse, 0, len(rows))
	for i := range rows {
		tools = append(
			tools,
			*responses.ToSelfToolResponse(&rows[i], descriptionByName[rows[i].Name], groupByName[rows[i].Name]),
		)
	}
	return &responses.SelfToolListResponse{Tools: tools}, nil
}
