package middlewares_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models/x402"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment/middlewares"
	"github.com/stretchr/testify/assert"
)

// x402GateRule builds an x402 rule with its settings serialized into Config.
func x402GateRule(price string, payTo string, patterns []string) models.ExternalPaymentRule {
	config, _ := json.Marshal(x402.X402Config{
		Network:             "eip155:80002",
		Asset:               "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:               payTo,
		AssetTransferMethod: x402.AssetTransferMethodEip3009,
		Eip712Name:          "USDC",
		Eip712Version:       "2",
		AddressPatterns:     patterns,
	})
	return models.ExternalPaymentRule{
		Method: "x402",
		Price:  price,
		Config: config,
	}
}

func x402GatePublicRule() models.ExternalPaymentRule {
	return x402GateRule("10000", x402TestPayToHex, []string{".*"})
}

func x402GateFreeRule(patterns ...string) models.ExternalPaymentRule {
	return x402GateRule("0", x402TestAgentWalletHex, patterns)
}

// x402GatePermit2Rule is a public paid rule whose token resolves to the
// permit2 transfer method — it carries no EIP-712 domain.
func x402GatePermit2Rule() models.ExternalPaymentRule {
	config, _ := json.Marshal(x402.X402Config{
		Network:             "eip155:80002",
		Asset:               "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:               x402TestPayToHex,
		AssetTransferMethod: x402.AssetTransferMethodPermit2,
		AddressPatterns:     []string{".*"},
	})
	return models.ExternalPaymentRule{Method: "x402", Price: "10000", Config: config}
}

func x402GateTestRules(rules ...models.ExternalPaymentRule) []models.ExternalPaymentRule {
	return rules
}

// x402GateFacilitators is the network → facilitator mapping for a node whose
// chain can settle. The "no facilitator" scenarios return an empty map.
func x402GateFacilitators() map[string]models.ExternalPaymentFacilitator {
	return map[string]models.ExternalPaymentFacilitator{
		"eip155:80002": {Network: "eip155:80002", BaseUrl: "https://x402-amoy.polygon.technology", MaxTimeoutSeconds: 120},
	}
}

// x402GatePaymentHeader builds a PAYMENT-SIGNATURE header aimed at `to` with
// the given value on the given network.
func x402GatePaymentHeader(t *testing.T, network string, to string, value string) string {
	t.Helper()

	header, err := x402.EncodeBase64Json(&x402.PaymentPayload{
		X402Version: x402.Version,
		Scheme:      x402.SchemeExact,
		Network:     network,
		Payload: x402.ExactEvmPayload{
			Signature: "0xsignature",
			Authorization: &x402.Eip3009Authorization{
				From:        x402TestPayerHex,
				To:          to,
				Value:       value,
				ValidAfter:  "0",
				ValidBefore: "9999999999",
				Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
	})
	assert.NoError(t, err)
	return header
}

func x402GateRequest(paymentHeader string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/responses", nil)
	req.Header.Set("Content-Type", "application/json")
	if len(paymentHeader) != 0 {
		req.Header.Set(middlewares.PaymentSignatureHeader, paymentHeader)
	}
	return req
}

func decodePaymentRequiredHeader(t *testing.T, resp *http.Response) *x402.PaymentRequired {
	t.Helper()

	raw := resp.Header.Get(middlewares.PaymentRequiredHeader)
	assert.NotEmpty(t, raw)
	decoded, err := base64.StdEncoding.DecodeString(raw)
	assert.NoError(t, err)
	var body x402.PaymentRequired
	assert.NoError(t, json.Unmarshal(decoded, &body))
	return &body
}
