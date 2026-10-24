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

type WhenCallingAmazonBedrockConverseTestingSuite struct {
	sut    *clients.AmazonBedrockConverseStandardClient
	server *httptest.Server
}

func WhenCallingAmazonBedrockConverseBeforeEach(
	t *testing.T,
	handler http.HandlerFunc,
) *WhenCallingAmazonBedrockConverseTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewAmazonBedrockConverseStandardClient(clients.NewAmazonBedrockConverseStandardClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingAmazonBedrockConverseTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingAmazonBedrockConverseAfterEach(suite *WhenCallingAmazonBedrockConverseTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingAmazonBedrockConverse(t *testing.T) {
	t.Parallel()

	t.Run("Given the upstream returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return a LlmStatusError with the 500", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAmazonBedrockConverseBeforeEach(t, handler)
			defer WhenCallingAmazonBedrockConverseAfterEach(suite)

			_, err := suite.sut.Messages(
				suite.server.URL,
				"sk",
				&models.ExternalAmazonBedrockConverseStandardRequest{ModelId: "x"},
			)
			var statusError *clients.LlmStatusError

			if assert.ErrorAs(t, err, &statusError) {
				assert.Equal(t, http.StatusInternalServerError, statusError.StatusCode)
			}
		})
	})

	t.Run("Given the upstream returns 200", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenAuth string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"output":{"message":{"role":"assistant","content":[{"text":"hi"}]}}}`))
		}

		t.Run("Should POST to /model/<modelId>/converse with bearer auth", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAmazonBedrockConverseBeforeEach(t, handler)
			defer WhenCallingAmazonBedrockConverseAfterEach(suite)

			res, err := suite.sut.Messages(
				suite.server.URL,
				"sk-secret",
				&models.ExternalAmazonBedrockConverseStandardRequest{
					ModelId: "anthropic.claude-3-sonnet",
				},
			)

			assert.NoError(t, err)
			assert.Equal(t, "/model/anthropic.claude-3-sonnet/converse", seenPath)
			assert.Equal(t, "Bearer sk-secret", seenAuth)
			assert.Equal(t, "assistant", res.Output.Message.Role)
		})
	})
}
