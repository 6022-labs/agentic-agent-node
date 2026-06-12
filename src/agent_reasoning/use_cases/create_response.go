package use_cases

import (
	"context"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
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
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

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
		return nil, errors.NewInternalError("create_response_failed", err.Error())
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
