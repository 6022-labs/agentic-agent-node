package clients_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenSendingA2AMessageTestingSuite struct {
	sut    *clients.A2AClient
	server *httptest.Server

	capturedMethod  string
	capturedPath    string
	capturedHeaders http.Header
	capturedBody    []byte
}

func WhenSendingA2AMessageBeforeEach(
	t *testing.T,
	maxResponseBytes int64,
	respond http.HandlerFunc,
) *WhenSendingA2AMessageTestingSuite {
	t.Helper()

	suite := &WhenSendingA2AMessageTestingSuite{
		sut: clients.NewA2AClient(clients.NewA2AClientParams{
			Logger:      zap.NewNop(),
			A2ASettings: &settings.A2ASettings{RequestTimeoutSeconds: 5, MaxResponseBytes: maxResponseBytes},
			HTTPClient:  http.DefaultClient,
		}),
	}
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.capturedMethod = r.Method
		suite.capturedPath = r.URL.Path
		suite.capturedHeaders = r.Header.Clone()
		suite.capturedBody, _ = io.ReadAll(r.Body)
		respond(w, r)
	}))
	return suite
}

func WhenSendingA2AMessageAfterEach(suite *WhenSendingA2AMessageTestingSuite) {
	suite.server.Close()
}

func (suite *WhenSendingA2AMessageTestingSuite) send(
	paymentSignature string,
) (*models.ExternalA2AResponse, *x402.PaymentRequired, error) {
	return suite.sut.SendMessage(context.Background(), suite.server.URL+"/api/a2a", &models.ExternalA2ARequest{
		JsonRpc: "2.0",
		Id:      "request-1",
		Method:  "SendMessage",
		Params: models.ExternalA2AMessageSendParams{
			Message: models.ExternalA2AMessage{
				MessageId: "message-1",
				Role:      "ROLE_USER",
				Parts:     []models.ExternalA2APart{{Text: "hello"}},
			},
		},
	}, paymentSignature)
}

func TestWhenSendingA2AMessage(t *testing.T) {
	t.Parallel()

	const maxResponseBytes = 1 << 20
	payTo := common_tests.RandomEthAddress().Hex()
	asset := common_tests.RandomEthAddress().Hex()

	t.Run("Given the agent answers the turn", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"request-1","result":{"message":{"messageId":"m2",` +
				`"role":"ROLE_AGENT","parts":[{"text":"hi "},{"text":"there"}],"contextId":"conversation-1"}}}`))
		}

		t.Run("Should post to the endpoint", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("eyJ4NDAy")

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, suite.capturedMethod)
			assert.Equal(t, "/api/a2a", suite.capturedPath)
		})

		t.Run("Should send the json-rpc envelope", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("eyJ4NDAy")

			assert.NoError(t, err)
			assert.Equal(t, "application/json", suite.capturedHeaders.Get("Content-Type"))
			assert.JSONEq(t, `{"jsonrpc":"2.0","id":"request-1","method":"SendMessage","params":{"message":`+
				`{"messageId":"message-1","role":"ROLE_USER","parts":[{"text":"hello"}]}}}`, string(suite.capturedBody))
		})

		t.Run("Should send the payment signature header", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("eyJ4NDAy")

			assert.NoError(t, err)
			assert.Equal(t, "eyJ4NDAy", suite.capturedHeaders.Get("PAYMENT-SIGNATURE"))
		})

		t.Run("Should join the reply parts", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			response, _, err := suite.send("eyJ4NDAy")

			assert.NoError(t, err)
			assert.Equal(t, "hi there", response.Text())
		})

		t.Run("Should return the reply's context id", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			response, _, err := suite.send("eyJ4NDAy")

			assert.NoError(t, err)
			assert.Equal(t, "conversation-1", response.ContextId())
		})
	})

	t.Run("Given no payment signature", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"message":{"parts":[{"text":"hi"}]}}}`))
		}

		t.Run("Should omit the payment header", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("")

			assert.NoError(t, err)
			assert.NotContains(t, suite.capturedHeaders, "Payment-Signature")
		})
	})

	t.Run("Given a response larger than the configured limit", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"message":{"parts":[{"text":"` +
				strings.Repeat("x", 512) + `"}]}}}`))
		}

		t.Run("Should report the overrun", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, 64, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("")

			assert.ErrorContains(t, err, "exceeds the 64 byte limit")
		})
	})

	t.Run("Given the agent quotes a price in the body", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`{"x402Version":2,"resource":{"url":"https://peer/api/a2a"},` +
				`"accepts":[{"scheme":"exact","network":"eip155:80002","amount":"10000",` +
				`"asset":"` + asset + `","payTo":"` + payTo + `"}]}`))
		}

		t.Run("Should return the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, paymentRequired, err := suite.send("")

			assert.NoError(t, err)
			assert.Equal(t, []x402.PaymentRequirements{{
				Scheme:  x402.SchemeExact,
				Network: "eip155:80002",
				Amount:  "10000",
				Asset:   asset,
				PayTo:   payTo,
			}}, paymentRequired.Accepts)
		})

		t.Run("Should return no reply", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			response, _, err := suite.send("")

			assert.NoError(t, err)
			assert.Nil(t, response)
		})
	})

	t.Run("Given the agent quotes a price in the header only", func(t *testing.T) {
		t.Parallel()

		quote, _ := json.Marshal(&x402.PaymentRequired{
			X402Version: x402.Version,
			Accepts:     []x402.PaymentRequirements{{Scheme: x402.SchemeExact, Amount: "500"}},
		})
		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("PAYMENT-REQUIRED", base64.StdEncoding.EncodeToString(quote))
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`payment required`))
		}

		t.Run("Should decode the quote from the header", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, paymentRequired, err := suite.send("")

			assert.NoError(t, err)
			assert.Equal(
				t,
				[]x402.PaymentRequirements{{Scheme: x402.SchemeExact, Amount: "500"}},
				paymentRequired.Accepts,
			)
		})
	})

	t.Run("Given a 402 that quotes nothing payable", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`{"x402Version":2,"accepts":[],"error":"no_matching_payment_requirements"}`))
		}

		t.Run("Should surface the peer's stated reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, paymentRequired, err := suite.send("")

			assert.NoError(t, err)
			assert.Equal(t, "no_matching_payment_requirements", paymentRequired.Error)
		})
	})

	t.Run("Given a 402 whose body and header disagree", func(t *testing.T) {
		t.Parallel()

		quote, _ := json.Marshal(&x402.PaymentRequired{
			X402Version: x402.Version,
			Accepts:     []x402.PaymentRequirements{{Amount: "999"}},
		})
		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("PAYMENT-REQUIRED", base64.StdEncoding.EncodeToString(quote))
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`{"x402Version":2,"error":"authorization_expired","accepts":[]}`))
		}

		t.Run("Should take the body alone", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, paymentRequired, err := suite.send("")

			assert.NoError(t, err)
			assert.Equal(t, &x402.PaymentRequired{
				X402Version: x402.Version,
				Error:       "authorization_expired",
				Accepts:     []x402.PaymentRequirements{},
			}, paymentRequired)
		})
	})

	t.Run("Given a 402 with no decodable quote", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusPaymentRequired)
			_, _ = w.Write([]byte(`payment required`))
		}

		t.Run("Should still report the payment as required", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, paymentRequired, err := suite.send("")

			assert.NoError(t, err)
			assert.Equal(t, &x402.PaymentRequired{X402Version: x402.Version}, paymentRequired)
		})
	})

	t.Run("Given the agent fails the request", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`runtime unavailable`))
		}

		t.Run("Should report the status", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("")

			assert.ErrorContains(t, err, "a2a request returned 502")
		})

		t.Run("Should withhold the body", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("")

			assert.NotContains(t, err.Error(), "runtime unavailable")
		})
	})

	t.Run("Given the agent answers with an undecodable reply", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not json`))
		}

		t.Run("Should report the decode failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t, maxResponseBytes, respond)
			defer WhenSendingA2AMessageAfterEach(suite)

			_, _, err := suite.send("")

			assert.ErrorContains(t, err, "failed to decode a2a response")
		})
	})
}
