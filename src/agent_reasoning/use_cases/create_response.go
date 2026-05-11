package use_cases

import (
	"context"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

type CreateResponseInterface interface {
	Execute(ctx context.Context, request *requests.CreateResponseRequest) (*responses.CreateResponseResponse, error)
}

type CreateResponse struct {
	responseDriver interfaces.ResponseDriverInterface
}

func NewCreateResponse(responseDriver interfaces.ResponseDriverInterface) *CreateResponse {
	return &CreateResponse{responseDriver: responseDriver}
}

func (u *CreateResponse) Execute(ctx context.Context, request *requests.CreateResponseRequest) (*responses.CreateResponseResponse, error) {
	inputs := make([]knowledge_interfaces.ConversationMessage, 0, len(request.Inputs))
	for _, m := range request.Inputs {
		inputs = append(inputs, knowledge_interfaces.ConversationMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	persisted, err := u.responseDriver.Drive(ctx, request.ConversationId, inputs)
	if err != nil {
		return nil, err
	}
	if persisted == nil {
		return &responses.CreateResponseResponse{}, nil
	}
	return &responses.CreateResponseResponse{
		Message: &responses.CreateResponseMessage{
			Role:      persisted.Role,
			Content:   persisted.Content,
			Timestamp: persisted.Timestamp,
		},
	}, nil
}
