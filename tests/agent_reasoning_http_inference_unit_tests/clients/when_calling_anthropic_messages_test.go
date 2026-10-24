package clients_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCallingAnthropicMessagesTestingSuite struct {
	sut    *clients.AnthropicMessagesStandardClient
	server *httptest.Server
}

func WhenCallingAnthropicMessagesBeforeEach(
	t *testing.T,
	handler http.HandlerFunc,
) *WhenCallingAnthropicMessagesTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewAnthropicMessagesStandardClient(clients.NewAnthropicMessagesStandardClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingAnthropicMessagesTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingAnthropicMessagesAfterEach(suite *WhenCallingAnthropicMessagesTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingAnthropicMessages(t *testing.T) {
	t.Parallel()

	t.Run("Given the upstream returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return a LlmStatusError with the 500", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAnthropicMessagesBeforeEach(t, handler)
			defer WhenCallingAnthropicMessagesAfterEach(suite)

			_, err := suite.sut.Messages(
				suite.server.URL,
				"sk",
				&models.ExternalAnthropicMessagesStandardRequest{Model: "claude"},
			)
			var statusError *clients.LlmStatusError

			if assert.ErrorAs(t, err, &statusError) {
				assert.Equal(t, http.StatusInternalServerError, statusError.StatusCode)
			}
		})
	})

	t.Run("Given the upstream returns 200", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenApiKey, seenVersion, seenBeta string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenApiKey = r.Header.Get("x-api-key")
			seenVersion = r.Header.Get("anthropic-version")
			seenBeta = r.Header.Get("anthropic-beta")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(
				[]byte(`{"id":"msg-1","model":"claude","role":"assistant","content":[{"type":"text","text":"hi"}]}`),
			)
		}

		t.Run("Should POST to /v1/messages with x-api-key + anthropic-version + anthropic-beta", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAnthropicMessagesBeforeEach(t, handler)
			defer WhenCallingAnthropicMessagesAfterEach(suite)

			res, err := suite.sut.Messages(
				suite.server.URL,
				"sk-secret",
				&models.ExternalAnthropicMessagesStandardRequest{Model: "claude"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "/v1/messages", seenPath)
			assert.Equal(t, "sk-secret", seenApiKey)
			assert.Equal(t, "2023-06-01", seenVersion)
			assert.NotEmpty(t, seenBeta)
			assert.Equal(t, "msg-1", res.ID)
		})
	})
}
