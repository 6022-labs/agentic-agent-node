package use_cases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type SendA2AMessageInterface interface {
	// Execute maps one A2A SendMessage turn onto the runtime Responses flow and back.
	Execute(ctx context.Context, request *requests.A2ARequest) (*responses.A2AResponse, error)
}

type SendA2AMessage struct {
	createResponse CreateResponseInterface
}

func NewSendA2AMessage(createResponse CreateResponseInterface) *SendA2AMessage {
	return &SendA2AMessage{createResponse: createResponse}
}

func (u *SendA2AMessage) Execute(ctx context.Context, request *requests.A2ARequest) (*responses.A2AResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(request.Params.Message.Text())
	createRequest := &requests.CreateResponseRequest{
		Input: requests.ResponseInput{
			Messages: []requests.InputMessage{{
				Role:    "user",
				Content: text,
			}},
		},
	}
	contextId := request.Params.Message.ContextId
	if contextId != "" {
		createRequest.Conversation = &requests.ConversationRef{
			Id: contextId,
		}
	}

	result, err := u.createResponse.Execute(ctx, createRequest)
	if err != nil {
		return nil, err
	}

	return &responses.A2AResponse{
		JsonRpc: "2.0",
		Id:      request.Id,
		Result: &responses.SendMessageResult{
			Message: &responses.A2AMessage{
				MessageId: uuid.NewString(),
				Role:      a2aRoleAgent,
				Parts:     []responses.A2APart{{Text: u.assistantText(result)}},
				ContextId: u.responseContextId(result, contextId),
			},
		},
	}, nil
}

const a2aRoleAgent = "ROLE_AGENT"

func (u *SendA2AMessage) assistantText(result *responses.Response) string {
	var builder strings.Builder
	for _, item := range result.Output {
		if item.Role != "" && item.Role != "assistant" {
			continue
		}
		for _, content := range item.Content {
			switch content.Type {
			case "", "text", "output_text":
				builder.WriteString(content.Text)
			}
		}
	}
	return builder.String()
}

func (u *SendA2AMessage) responseContextId(result *responses.Response, fallback string) string {
	if result.Conversation != nil && strings.TrimSpace(result.Conversation.Id) != "" {
		return result.Conversation.Id
	}
	return fallback
}
