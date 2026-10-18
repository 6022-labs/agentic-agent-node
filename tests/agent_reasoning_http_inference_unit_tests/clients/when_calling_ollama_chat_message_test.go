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

type WhenCallingOllamaChatMessageTestingSuite struct {
	sut    *clients.OllamaChatMessageStandardClient
	server *httptest.Server
}

func WhenCallingOllamaChatMessageBeforeEach(
	t *testing.T,
	handler http.HandlerFunc,
) *WhenCallingOllamaChatMessageTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewOllamaChatMessageStandardClient(clients.NewOllamaChatMessageStandardClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingOllamaChatMessageTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingOllamaChatMessageAfterEach(suite *WhenCallingOllamaChatMessageTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingOllamaChatMessage(t *testing.T) {
	t.Parallel()

	t.Run("Given the upstream returns 200", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenAuth string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"model":"llama3","message":{"role":"assistant","content":"hi"}}`))
		}

		t.Run("Should return the decoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOllamaChatMessageBeforeEach(t, handler)
			defer WhenCallingOllamaChatMessageAfterEach(suite)

			res, err := suite.sut.ChatMessage(
				suite.server.URL,
				&models.ExternalOllamaChatMessageStandardRequest{Model: "llama3"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "/api/chat", seenPath)
			assert.Empty(t, seenAuth)
			assert.Equal(t, "llama3", res.Model)
			assert.Equal(t, "hi", res.Message.Content)
		})
	})

	t.Run("Given the upstream returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return a LlmStatusError with the 500", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOllamaChatMessageBeforeEach(t, handler)
			defer WhenCallingOllamaChatMessageAfterEach(suite)

			_, err := suite.sut.ChatMessage(
				suite.server.URL,
				&models.ExternalOllamaChatMessageStandardRequest{Model: "llama3"},
			)
			var statusError *clients.LlmStatusError

			if assert.ErrorAs(t, err, &statusError) {
				assert.Equal(t, http.StatusInternalServerError, statusError.StatusCode)
			}
		})
	})
}
