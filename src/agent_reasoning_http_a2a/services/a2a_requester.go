package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/clients"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/google/uuid"
)

// a2aMethodSendMessage is the only method 6022 agents answer; the "message/send" spelling is rejected.
const a2aMethodSendMessage = "SendMessage"

const a2aRoleUser = "ROLE_USER"

type A2ARequester struct {
	a2aClient  clients.A2AClientInterface
	a2aMetrics metrics_interfaces.A2AMetricsInterface
}

func NewA2ARequester(
	a2aClient clients.A2AClientInterface,
	a2aMetrics metrics_interfaces.A2AMetricsInterface,
) *A2ARequester {
	return &A2ARequester{a2aClient: a2aClient, a2aMetrics: a2aMetrics}
}

func (r *A2ARequester) FetchAgentCard(
	ctx context.Context,
	baseUrl string,
) (*reasoning_interfaces.A2AAgentCard, error) {
	start := time.Now()
	card, err := r.a2aClient.FetchAgentCard(ctx, baseUrl)
	r.a2aMetrics.RecordFetchAgentCard(ctx, baseUrl, time.Since(start), err)
	if err != nil {
		return nil, err
	}

	return &reasoning_interfaces.A2AAgentCard{
		Name:        card.Name,
		EndpointUrl: card.JsonRpcUrl(),
	}, nil
}

func (r *A2ARequester) SendMessage(
	ctx context.Context,
	endpointUrl string,
	message string,
	contextId string,
	paymentSignature string,
) (*reasoning_interfaces.A2AReply, *x402.PaymentRequired, error) {
	request := &models.ExternalA2ARequest{
		JsonRpc: "2.0",
		Id:      uuid.NewString(),
		Method:  a2aMethodSendMessage,
		Params: models.ExternalA2AMessageSendParams{
			Message: models.ExternalA2AMessage{
				MessageId: uuid.NewString(),
				Role:      a2aRoleUser,
				Parts:     []models.ExternalA2APart{{Text: message}},
				ContextId: contextId,
			},
		},
	}

	start := time.Now()
	response, paymentRequired, err := r.a2aClient.SendMessage(ctx, endpointUrl, request, paymentSignature)
	r.a2aMetrics.RecordSendMessage(ctx, endpointUrl, time.Since(start), err, paymentRequired != nil)
	if err != nil {
		return nil, nil, err
	}
	if paymentRequired != nil {
		return nil, paymentRequired, nil
	}
	if response == nil {
		return nil, nil, errors.New("agent returned neither a reply nor a payment quote")
	}
	if response.Error != nil {
		return nil, nil, fmt.Errorf("agent returned error %d: %s", response.Error.Code, response.Error.Message)
	}
	// A task or any other result shape is unsupported rather than read as an empty reply.
	if response.Result == nil || response.Result.Message == nil {
		return nil, nil, errors.New("agent replied without a message")
	}

	return &reasoning_interfaces.A2AReply{
		Text:      response.Text(),
		ContextId: response.ContextId(),
	}, nil, nil
}
