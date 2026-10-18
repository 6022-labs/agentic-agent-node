package use_cases

import (
	"context"
	goerrors "errors"
	"time"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/google/uuid"
)

type CreateResponseInterface interface {
	// Execute runs the turn in the caller's session and echoes the key it sent, never the scoped storage id.
	Execute(ctx context.Context, request *requests.CreateResponseRequest) (*responses.CreateResponseResponse, error)
}

type CreateResponse struct {
	responseDriver       interfaces.ResponseDriverInterface
	conversationResolver knowledge_interfaces.ConversationResolverInterface
	responseMetrics      metrics_interfaces.ResponseMetricsInterface
	reasoningTracer      traces_interfaces.ReasoningTracerInterface
}

func NewCreateResponse(
	responseDriver interfaces.ResponseDriverInterface,
	conversationResolver knowledge_interfaces.ConversationResolverInterface,
	responseMetrics metrics_interfaces.ResponseMetricsInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
) *CreateResponse {
	return &CreateResponse{
		responseDriver:       responseDriver,
		conversationResolver: conversationResolver,
		responseMetrics:      responseMetrics,
		reasoningTracer:      reasoningTracer,
	}
}

func (u *CreateResponse) Execute(
	ctx context.Context,
	request *requests.CreateResponseRequest,
) (*responses.CreateResponseResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	conversationId, err := u.conversationResolver.Resolve(ctx, request.ConversationKey(), callerId)
	if err != nil {
		u.responseMetrics.RecordResponseFailed(ctx)
		return nil, u.resolveError(err)
	}
	echoedId := request.ConversationKey()
	if echoedId == "" {
		echoedId = conversationId
	}

	ctx, span := u.reasoningTracer.StartInference(ctx, conversationId, request.Model)
	defer span.End()

	inputs := make([]knowledge_interfaces.ConversationMessage, 0, len(request.Input))
	for _, message := range request.Input {
		inputs = append(inputs, knowledge_interfaces.ConversationMessage{
			Role:      message.Role,
			Content:   message.Content,
			Timestamp: time.Now(),
		})
	}

	persisted, err := u.responseDriver.Drive(ctx, conversationId, request.Instructions, inputs)
	if err != nil {
		span.Fail(err)
		u.responseMetrics.RecordResponseFailed(ctx)
		return nil, errors.NewInternalError("create_response_failed", err.Error())
	}

	u.responseMetrics.RecordResponseSucceeded(ctx)
	return u.buildResponse(request, echoedId, persisted), nil
}

// resolveError maps caller-caused outcomes to 409; anything else is a 500.
func (u *CreateResponse) resolveError(err error) error {
	switch {
	case goerrors.Is(err, knowledge_interfaces.ErrConversationUnavailable):
		return errors.NewConflictError("conversation_unavailable", err.Error())
	case goerrors.Is(err, knowledge_interfaces.ErrAgentNotMinted):
		return errors.NewConflictError("agent_not_minted", err.Error())
	default:
		return errors.NewInternalError("create_response_failed", err.Error())
	}
}

func (u *CreateResponse) buildResponse(
	request *requests.CreateResponseRequest,
	conversationId string,
	persisted *knowledge_interfaces.ConversationMessage,
) *responses.CreateResponseResponse {
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
