package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type CreateResponseInterface interface {
	Execute(ctx context.Context, request *requests.CreateResponseRequest) *usecase_result.Result
}

type CreateResponse struct {
	client interfaces.ResponseClientInterface
}

func NewCreateResponse(client interfaces.ResponseClientInterface) *CreateResponse {
	return &CreateResponse{client: client}
}

// Execute translates the gateway request to the runtime contract. Prompt,
// Tools, Options, and ForceConclusion are swarm-side concerns — the
// orchestrator owns them. Agent-node only forwards inputs.
func (u *CreateResponse) Execute(ctx context.Context, request *requests.CreateResponseRequest) *usecase_result.Result {
	payload := &runtime_models.ExternalCreateResponsePayload{
		ConversationId: request.ConversationId,
		Inputs:         make([]runtime_models.ExternalCreateResponseInputPayload, 0, len(request.Input)),
	}
	for _, m := range request.Input {
		payload.Inputs = append(payload.Inputs, runtime_models.ExternalCreateResponseInputPayload{
			Role:      string(m.Role),
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	result, err := u.client.Create(ctx, payload)
	if err != nil {
		return usecase_result.Unavailable("runtime_unavailable", err.Error())
	}

	out := &responses.Response{
		IsFinal: result.Message != nil,
	}
	if result.Message != nil {
		out.Output = []*responses.OutputMessage{
			{
				Content:   result.Message.Content,
				Timestamp: result.Message.Timestamp,
			},
		}
	}
	return usecase_result.Ok(out)
}
