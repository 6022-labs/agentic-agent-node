package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/stretchr/testify/assert"
)

type WhenCreatingResponseOverHttpTestingSuite struct {
	sut    *services.HttpResponseClient
	server *httptest.Server
}

func WhenCreatingResponseOverHttpBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCreatingResponseOverHttpTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return &WhenCreatingResponseOverHttpTestingSuite{
		sut:    services.NewHttpResponseClient(client),
		server: server,
	}
}

func WhenCreatingResponseOverHttpAfterEach(suite *WhenCreatingResponseOverHttpTestingSuite) {
	suite.server.Close()
}

func TestWhenCreatingResponseOverHttp(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given the runtime returns an assistant message", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenMethod, seenBody string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			seenMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			seenBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"hi","timestamp":"` + now.Format(time.RFC3339Nano) + `"}}`))
		}

		t.Run("Should POST to /internal/responses and decode the result", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			result, err := suite.sut.Create(context.Background(), &models.CreateResponsePayload{
				ConversationId: "conv-1",
				Inputs: []models.CreateResponseInputPayload{
					{Role: "user", Content: "hello", Timestamp: now},
				},
			})

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, seenMethod)
			assert.Equal(t, "/internal/responses", seenPath)
			assert.Contains(t, seenBody, `"conversationId":"conv-1"`)
			assert.Contains(t, seenBody, `"role":"user"`)

			if assert.NotNil(t, result) && assert.NotNil(t, result.Message) {
				assert.Equal(t, "assistant", result.Message.Role)
				assert.Equal(t, "hi", result.Message.Content)
			}
		})
	})

	t.Run("Given the runtime returns an empty message", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}

		t.Run("Should return a result with nil message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			result, err := suite.sut.Create(context.Background(), &models.CreateResponsePayload{
				ConversationId: "conv-1",
				Inputs:         []models.CreateResponseInputPayload{{Role: "user", Content: "x", Timestamp: now}},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.Message)
			}
		})
	})

	t.Run("Given the runtime returns 500", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "create_response_failed"})
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			_, err := suite.sut.Create(context.Background(), &models.CreateResponsePayload{
				ConversationId: "conv-1",
				Inputs:         []models.CreateResponseInputPayload{{Role: "user", Content: "x", Timestamp: now}},
			})

			assert.Error(t, err)
		})
	})
}
