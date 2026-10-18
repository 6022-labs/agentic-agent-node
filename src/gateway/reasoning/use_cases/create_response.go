package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type CreateResponseInterface interface {
	Execute(ctx context.Context, request *requests.CreateResponseRequest) (*responses.Response, error)
}

type CreateResponse struct {
	client interfaces.ResponseClientInterface
}

func NewCreateResponse(client interfaces.ResponseClientInterface) *CreateResponse {
	return &CreateResponse{client: client}
}

// Execute authenticates (upstream, via the payment gate) then forwards the
// OpenAI Responses request to the runtime, which owns history and builds the
// response. The gateway only translates between its DTOs and the wire payload.
func (u *CreateResponse) Execute(
	ctx context.Context,
	request *requests.CreateResponseRequest,
) (*responses.Response, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	result, err := u.client.Create(ctx, toExternalPayload(request))
	if err != nil {
		return nil, err
	}

	return toResponse(result), nil
}

func toExternalPayload(request *requests.CreateResponseRequest) *models.ExternalCreateResponsePayload {
	inputs := make([]models.ExternalInputMessagePayload, 0, len(request.Input.Messages))
	for _, message := range request.Input.Messages {
		inputs = append(inputs, models.ExternalInputMessagePayload{
			Type:    "message",
			Role:    message.Role,
			Content: message.Content,
		})
	}

	payload := &models.ExternalCreateResponsePayload{
		Model:              request.Model,
		Instructions:       request.Instructions,
		Input:              inputs,
		PreviousResponseId: request.PreviousResponseId,
		Temperature:        request.Temperature,
		TopP:               request.TopP,
		MaxOutputTokens:    request.MaxOutputTokens,
		Store:              request.Store,
		Stream:             request.Stream,
		Metadata:           request.Metadata,
	}
	if request.Conversation != nil {
		payload.Conversation = &models.ExternalConversationRefPayload{Id: request.Conversation.Id}
	}
	return payload
}

func toResponse(result *models.ExternalCreateResponseResultPayload) *responses.Response {
	output := make([]responses.OutputItem, 0, len(result.Output))
	for _, item := range result.Output {
		content := make([]responses.OutputContent, 0, len(item.Content))
		for _, part := range item.Content {
			content = append(content, responses.OutputContent{
				Type:        part.Type,
				Text:        part.Text,
				Annotations: part.Annotations,
			})
		}
		output = append(output, responses.OutputItem{
			Type:    item.Type,
			Id:      item.Id,
			Status:  item.Status,
			Role:    item.Role,
			Content: content,
		})
	}

	response := &responses.Response{
		Id:                 result.Id,
		Object:             result.Object,
		CreatedAt:          result.CreatedAt,
		Status:             result.Status,
		Model:              result.Model,
		Output:             output,
		PreviousResponseId: result.PreviousResponseId,
		Metadata:           result.Metadata,
	}
	if result.Usage != nil {
		response.Usage = &responses.Usage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
			TotalTokens:  result.Usage.TotalTokens,
		}
	}
	if result.Conversation != nil {
		response.Conversation = &responses.ConversationRef{
			Id: result.Conversation.Id,
		}
	}
	if result.Error != nil {
		response.Error = &responses.ResponseError{
			Code:    result.Error.Code,
			Message: result.Error.Message,
		}
	}
	return response
}
