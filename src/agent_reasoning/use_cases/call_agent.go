package use_cases

import (
	"context"
	"net/url"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
)

// defaultA2APath is where a 6022 agent answers behind its nginx /api prefix when its card is unusable.
const defaultA2APath = "/api/a2a"

// signableExtraKeys are the 402 extra fields SignX402PaymentRequest reads.
var signableExtraKeys = []string{"assetTransferMethod", "name", "version"}

// CallAgent runs one A2A turn; a priced peer answers with a quote, and paying is a separate decision.
type CallAgent struct {
	a2aRequester interfaces.A2ARequesterInterface
}

func NewCallAgent(a2aRequester interfaces.A2ARequesterInterface) *CallAgent {
	return &CallAgent{a2aRequester: a2aRequester}
}

func (u *CallAgent) Execute(
	ctx context.Context,
	request *requests.CallAgentRequest,
) (*responses.CallAgentResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	endpointUrl, agentName := u.resolveEndpoint(ctx, request.Url)

	reply, paymentRequired, err := u.a2aRequester.SendMessage(
		ctx,
		endpointUrl,
		request.Message,
		request.ContextId,
		request.PaymentSignature,
	)
	if err != nil {
		return nil, errors.NewUnavailableError("agent_call_failed", "the agent could not be reached", err)
	}

	if paymentRequired != nil {
		return &responses.CallAgentResponse{
			Status:          responses.CallAgentStatusPaymentRequired,
			EndpointUrl:     endpointUrl,
			AgentName:       agentName,
			PaymentRequired: u.toPaymentRequiredResponse(paymentRequired),
		}, nil
	}
	if reply == nil {
		return nil, errors.NewUnavailableError("agent_call_failed", "the agent returned no reply", nil)
	}

	return &responses.CallAgentResponse{
		Status:      responses.CallAgentStatusOk,
		EndpointUrl: endpointUrl,
		AgentName:   agentName,
		Text:        reply.Text,
		ContextId:   reply.ContextId,
	}, nil
}

// resolveEndpoint prefers the card's JSONRPC url; a peer whose card is unusable still answers at the default path.
func (u *CallAgent) resolveEndpoint(ctx context.Context, baseUrl string) (endpointUrl string, agentName string) {
	card, err := u.a2aRequester.FetchAgentCard(ctx, baseUrl)
	if err != nil || card == nil {
		return baseUrl + defaultA2APath, ""
	}
	// The payment header travels on this request, so a card may move the endpoint within its origin only.
	if len(strings.TrimSpace(card.EndpointUrl)) == 0 || !u.sameOrigin(baseUrl, card.EndpointUrl) {
		return baseUrl + defaultA2APath, card.Name
	}
	return card.EndpointUrl, card.Name
}

func (u *CallAgent) sameOrigin(baseUrl string, endpointUrl string) bool {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return false
	}
	endpoint, err := url.Parse(endpointUrl)
	if err != nil || endpoint.User != nil {
		return false
	}
	return u.origin(base) == u.origin(endpoint)
}

// origin spells the default port out, so https://peer and https://peer:443 compare equal.
func (u *CallAgent) origin(parsed *url.URL) string {
	scheme := strings.ToLower(parsed.Scheme)
	port := parsed.Port()
	if len(port) == 0 {
		port = map[string]string{"http": "80", "https": "443"}[scheme]
	}
	return scheme + "://" + strings.ToLower(parsed.Hostname()) + ":" + port
}

// signableExtra keeps the string keys sign_x402_payment reads; the rest is peer data the model has no use for.
func (u *CallAgent) signableExtra(extra map[string]any) map[string]any {
	kept := make(map[string]any, len(signableExtraKeys))
	for _, key := range signableExtraKeys {
		if value, ok := extra[key].(string); ok {
			kept[key] = value
		}
	}
	if len(kept) == 0 {
		return nil
	}
	return kept
}

func (u *CallAgent) toPaymentRequiredResponse(
	paymentRequired *x402.PaymentRequired,
) *responses.PaymentRequiredResponse {
	accepts := make([]responses.PaymentRequirementResponse, 0, len(paymentRequired.Accepts))
	for _, accept := range paymentRequired.Accepts {
		accepts = append(accepts, responses.PaymentRequirementResponse{
			Scheme:            accept.Scheme,
			Network:           accept.Network,
			Amount:            accept.Amount,
			Asset:             accept.Asset,
			PayTo:             accept.PayTo,
			MaxTimeoutSeconds: accept.MaxTimeoutSeconds,
			Extra:             u.signableExtra(accept.Extra),
		})
	}
	return &responses.PaymentRequiredResponse{
		X402Version: paymentRequired.X402Version,
		Resource:    paymentRequired.Resource.Url,
		Error:       paymentRequired.Error,
		Accepts:     accepts,
	}
}
