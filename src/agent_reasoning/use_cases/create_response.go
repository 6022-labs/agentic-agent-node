package use_cases

import (
	"context"
	"time"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
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

// Execute resolves the conversation key (minting one when absent), forwards the
// new turn to the inference loop — instructions ride along as a leading system
// message — and builds the OpenAI Response.
func (u *CreateResponse) Execute(ctx context.Context, request *requests.CreateResponseRequest) (*responses.CreateResponseResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	conversationId := request.ConversationKey()
	if conversationId == "" {
		conversationId = uuid.NewString()
	}

	inputs := make([]knowledge_interfaces.ConversationMessage, 0, len(request.Input)+1)
	if request.Instructions != "" {
		inputs = append(inputs, knowledge_interfaces.ConversationMessage{
			Role:      "system",
			Content:   request.Instructions,
			Timestamp: time.Now(),
		})
	}
	for _, message := range request.Input {
		inputs = append(inputs, knowledge_interfaces.ConversationMessage{
			Role:      message.Role,
			Content:   message.Content,
			Timestamp: time.Now(),
		})
	}

	persisted, err := u.responseDriver.Drive(ctx, conversationId, inputs)
	if err != nil {
		return nil, errors.NewInternalError("create_response_failed", err.Error())
	}

	return buildResponse(request, conversationId, persisted), nil
}

func buildResponse(request *requests.CreateResponseRequest, conversationId string, persisted *knowledge_interfaces.ConversationMessage) *responses.CreateResponseResponse {
	response := &responses.CreateResponseResponse{
		Id:           conversationId,
		Object:       "response",
		CreatedAt:    time.Now().Unix(),
		Status:       "incomplete",
		Model:        request.Model,
		Output:       []responses.OutputItem{},
		Conversation: &responses.ConversationRef{Id: conversationId},
		Metadata:     request.Metadata,
	}
	if previous := request.PreviousResponseId; previous != "" {
		response.PreviousResponseId = &previous
	}

	if persisted != nil {
		response.Status = "completed"
		response.Output = []responses.OutputItem{
			{
				Type:   "message",
				Id:     "msg_" + uuid.NewString(),
				Status: "completed",
				Role:   "assistant",
				Content: []responses.OutputContent{
					{
						Type:        "output_text",
						Text:        persisted.Content,
						Annotations: []any{},
					},
				},
			},
		}
	}

	return response
}
