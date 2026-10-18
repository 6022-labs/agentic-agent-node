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

type WhenCallingOpenAICompletionsTestingSuite struct {
	sut    *clients.OpenAICompletionsStandardClient
	server *httptest.Server
}

func WhenCallingOpenAICompletionsBeforeEach(
	t *testing.T,
	handler http.HandlerFunc,
) *WhenCallingOpenAICompletionsTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewOpenAICompletionsStandardClient(clients.NewOpenAICompletionsStandardClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingOpenAICompletionsTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingOpenAICompletionsAfterEach(suite *WhenCallingOpenAICompletionsTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingOpenAICompletions(t *testing.T) {
	t.Parallel()

	t.Run("Given the upstream returns 200", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenAuth string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(
				[]byte(
					`{"id":"chatcmpl-1","object":"chat.completion","model":"gpt-4o","choices":[{"index":0,` +
						`"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`,
				),
			)
		}

		t.Run("Should return the decoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOpenAICompletionsBeforeEach(t, handler)
			defer WhenCallingOpenAICompletionsAfterEach(suite)

			res, err := suite.sut.ChatCompletion(
				suite.server.URL,
				"sk-secret",
				&models.ExternalOpenAICompletionsStandardRequest{Model: "gpt-4o"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "/v1/chat/completions", seenPath)
			assert.Equal(t, "Bearer sk-secret", seenAuth)
			assert.Equal(t, "chatcmpl-1", res.ID)
			assert.Len(t, res.Choices, 1)
			assert.Equal(t, "hi", res.Choices[0].Message.Content)
		})
	})

	t.Run("Given the upstream returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return a LlmStatusError with the 500", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOpenAICompletionsBeforeEach(t, handler)
			defer WhenCallingOpenAICompletionsAfterEach(suite)

			_, err := suite.sut.ChatCompletion(
				suite.server.URL,
				"sk",
				&models.ExternalOpenAICompletionsStandardRequest{Model: "gpt-4o"},
			)
			var statusError *clients.LlmStatusError

			if assert.ErrorAs(t, err, &statusError) {
				assert.Equal(t, http.StatusInternalServerError, statusError.StatusCode)
			}
		})
	})
}
