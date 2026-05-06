package clients_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCallingOpenAIResponsesTestingSuite struct {
	sut    *clients.OpenAIResponsesStandardClient
	server *httptest.Server
}

func WhenCallingOpenAIResponsesBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCallingOpenAIResponsesTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewOpenAIResponsesStandardClient(clients.NewOpenAIResponsesStandardClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingOpenAIResponsesTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingOpenAIResponsesAfterEach(suite *WhenCallingOpenAIResponsesTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingOpenAIResponses(t *testing.T) {
	t.Parallel()

	t.Run("Given the upstream returns 200", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenAuth string
		var seenBody []byte
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenAuth = r.Header.Get("Authorization")
			seenBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"resp-1","model":"gpt-4o","output":[{"type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"hello"}]}]}`))
		}

		t.Run("Should POST to /v1/responses with bearer auth and decode the response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOpenAIResponsesBeforeEach(t, handler)
			defer WhenCallingOpenAIResponsesAfterEach(suite)

			res, err := suite.sut.Responses(suite.server.URL, "sk-secret", &models.ExternalOpenAIResponsesStandardRequest{
				Model: "gpt-4o",
			})

			assert.NoError(t, err)
			assert.Equal(t, "/v1/responses", seenPath)
			assert.Equal(t, "Bearer sk-secret", seenAuth)
			assert.Contains(t, string(seenBody), `"model":"gpt-4o"`)
			assert.Equal(t, "resp-1", res.ID)
		})
	})

	t.Run("Given the upstream returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOpenAIResponsesBeforeEach(t, handler)
			defer WhenCallingOpenAIResponsesAfterEach(suite)

			_, err := suite.sut.Responses(suite.server.URL, "sk", &models.ExternalOpenAIResponsesStandardRequest{Model: "gpt-4o"})
			assert.Error(t, err)
		})
	})

	t.Run("Given the upstream returns malformed JSON", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not json`))
		}

		t.Run("Should return a decode error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingOpenAIResponsesBeforeEach(t, handler)
			defer WhenCallingOpenAIResponsesAfterEach(suite)

			_, err := suite.sut.Responses(suite.server.URL, "sk", &models.ExternalOpenAIResponsesStandardRequest{Model: "gpt-4o"})
			assert.ErrorContains(t, err, "decode")
		})
	})
}

