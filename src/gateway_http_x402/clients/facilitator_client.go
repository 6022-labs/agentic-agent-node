package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models/x402"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const (
	FacilitatorVerifyPath = "/verify"
	FacilitatorSettlePath = "/settle"
)

// facilitatorRequest is the body shape of POST /verify and POST /settle.
type facilitatorRequest struct {
	X402Version         int                       `json:"x402Version"`
	PaymentPayload      *x402.PaymentPayload      `json:"paymentPayload"`
	PaymentRequirements *x402.PaymentRequirements `json:"paymentRequirements"`
}

type FacilitatorClient struct {
	logger     *zap.Logger
	httpClient *http.Client
}

type NewFacilitatorClientParams struct {
	dig.In

	Logger     *zap.Logger
	HTTPClient *http.Client `name:"x402-facilitator-http-client"`
}

func NewFacilitatorClient(params NewFacilitatorClientParams) *FacilitatorClient {
	return &FacilitatorClient{
		logger:     params.Logger,
		httpClient: params.HTTPClient,
	}
}

func (c *FacilitatorClient) Verify(ctx context.Context, baseUrl string, payload *x402.PaymentPayload, requirements *x402.PaymentRequirements) (*x402.VerifyResponse, error) {
	var verifyResponse x402.VerifyResponse
	if err := c.post(ctx, baseUrl, FacilitatorVerifyPath, payload, requirements, &verifyResponse); err != nil {
		return nil, err
	}
	return &verifyResponse, nil
}

func (c *FacilitatorClient) Settle(ctx context.Context, baseUrl string, payload *x402.PaymentPayload, requirements *x402.PaymentRequirements) (*x402.SettleResponse, error) {
	var settleResponse x402.SettleResponse
	if err := c.post(ctx, baseUrl, FacilitatorSettlePath, payload, requirements, &settleResponse); err != nil {
		return nil, err
	}
	return &settleResponse, nil
}

func (c *FacilitatorClient) post(ctx context.Context, baseUrl string, path string, payload *x402.PaymentPayload, requirements *x402.PaymentRequirements, out any) error {
	encoded, err := json.Marshal(&facilitatorRequest{
		X402Version:         x402.Version,
		PaymentPayload:      payload,
		PaymentRequirements: requirements,
	})
	if err != nil {
		return fmt.Errorf("failed to encode facilitator request: %w", err)
	}

	url := strings.TrimRight(baseUrl, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("facilitator returned %d (%s)", resp.StatusCode, string(bodyBytes))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode facilitator response: %w", err)
	}
	return nil
}
