package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const (
	// AgentCardPath is the A2A v1.0 discovery document, served unauthenticated.
	AgentCardPath = "/.well-known/agent-card.json"
	// PaymentSignatureHeader carries the x402 v2 payload.
	PaymentSignatureHeader = "PAYMENT-SIGNATURE"
	// PaymentRequiredHeader carries the base64 402 quote.
	PaymentRequiredHeader = "PAYMENT-REQUIRED"
)

type A2AClientInterface interface {
	// FetchAgentCard reads the peer's agent card; a peer that serves none is not an error to the caller.
	FetchAgentCard(ctx context.Context, baseUrl string) (*models.ExternalAgentCard, error)
	// SendMessage posts one JSON-RPC SendMessage. A 402 comes back as a quote with a nil response.
	SendMessage(
		ctx context.Context,
		endpointUrl string,
		request *models.ExternalA2ARequest,
		paymentSignature string,
	) (*models.ExternalA2AResponse, *x402.PaymentRequired, error)
}

type A2AClient struct {
	logger      *zap.Logger
	httpClient  *http.Client
	a2aSettings *settings.A2ASettings
}

type NewA2AClientParams struct {
	dig.In

	Logger      *zap.Logger
	A2ASettings *settings.A2ASettings
	HTTPClient  *http.Client `name:"a2a-http-client"`
}

func NewA2AClient(params NewA2AClientParams) *A2AClient {
	return &A2AClient{
		logger:      params.Logger,
		httpClient:  params.HTTPClient,
		a2aSettings: params.A2ASettings,
	}
}

func (client *A2AClient) FetchAgentCard(ctx context.Context, baseUrl string) (*models.ExternalAgentCard, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(client.a2aSettings.CardTimeoutSeconds)*time.Second)
	defer cancel()

	url := strings.TrimRight(strings.TrimSpace(baseUrl), "/") + AgentCardPath
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent card request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent card: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := client.readBody(response)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// Body withheld: the url can be model-chosen, so echoing it would make the fetch a read primitive.
		return nil, fmt.Errorf("agent card request returned %d", response.StatusCode)
	}

	var card models.ExternalAgentCard
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, fmt.Errorf("failed to decode agent card: %w", err)
	}
	return &card, nil
}

func (client *A2AClient) SendMessage(
	ctx context.Context,
	endpointUrl string,
	a2aRequest *models.ExternalA2ARequest,
	paymentSignature string,
) (*models.ExternalA2AResponse, *x402.PaymentRequired, error) {
	payload, err := json.Marshal(a2aRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode a2a request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build a2a request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if len(strings.TrimSpace(paymentSignature)) != 0 {
		request.Header.Set(PaymentSignatureHeader, paymentSignature)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call agent: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := client.readBody(response)
	if err != nil {
		return nil, nil, err
	}

	if response.StatusCode == http.StatusPaymentRequired {
		return nil, client.decodePaymentRequired(response, body), nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// Body withheld for the same reason as the card's.
		return nil, nil, fmt.Errorf("a2a request returned %d", response.StatusCode)
	}

	var a2aResponse models.ExternalA2AResponse
	if err := json.Unmarshal(body, &a2aResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to decode a2a response: %w", err)
	}
	return &a2aResponse, nil, nil
}

// decodePaymentRequired prefers the body over the header; an undecodable quote still reports a 402.
func (client *A2AClient) decodePaymentRequired(response *http.Response, body []byte) *x402.PaymentRequired {
	// Gated on decoding, not on accepts: a refusal carries empty accepts plus the reason the caller needs.
	var bodyQuote x402.PaymentRequired
	if err := json.Unmarshal(body, &bodyQuote); err == nil && bodyQuote.X402Version != 0 {
		return &bodyQuote
	}

	header := strings.TrimSpace(response.Header.Get(PaymentRequiredHeader))
	if len(header) != 0 {
		decoded, err := base64.StdEncoding.DecodeString(header)
		var headerQuote x402.PaymentRequired
		if err == nil && json.Unmarshal(decoded, &headerQuote) == nil {
			return &headerQuote
		}
	}

	client.logger.Warn("peer returned 402 with an undecodable payment quote")
	return &x402.PaymentRequired{X402Version: x402.Version}
}

// readBody reads one byte past the cap so an oversize response is named, not surfaced as a decode failure.
func (client *A2AClient) readBody(response *http.Response) ([]byte, error) {
	limit := client.a2aSettings.MaxResponseBytes
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	// No drain past the cap: a peer streaming forever would pin the call until its timeout.
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response exceeds the %d byte limit", limit)
	}
	return body, nil
}
