package services_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/stretchr/testify/assert"
)

type responseCaptured struct {
	method string
	path   string
	body   string
}

type WhenCreatingResponseOverHttpTestingSuite struct {
	sut      *services.HttpResponseClient
	server   *httptest.Server
	captured *responseCaptured
}

func WhenCreatingResponseOverHttpBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCreatingResponseOverHttpTestingSuite {
	t.Helper()
	captured := &responseCaptured{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		captured.body = string(body)
		handler(w, r)
	}))
	client, _ := clients.NewClientWithResponses(server.URL)
	return &WhenCreatingResponseOverHttpTestingSuite{
		sut:      services.NewHttpResponseClient(client),
		server:   server,
		captured: captured,
	}
}

func WhenCreatingResponseOverHttpAfterEach(suite *WhenCreatingResponseOverHttpTestingSuite) {
	suite.server.Close()
}

func TestWhenCreatingResponseOverHttp(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	request := &models.ExternalCreateResponsePayload{
		ConversationId: "conv-1",
		Inputs: []models.ExternalCreateResponseInputPayload{
			{Role: "user", Content: "hello", Timestamp: now},
		},
	}

	t.Run("Given the runtime returns an assistant message", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"hi","timestamp":"2026-01-01T00:00:00Z"}}`))
		}

		t.Run("Should POST to /internal/responses and decode the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, suite.captured.method)
			assert.Equal(t, "/internal/responses", suite.captured.path)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Message) {
				assert.Equal(t, "assistant", result.Message.Role)
				assert.Equal(t, "hi", result.Message.Content)
			}
		})
	})

	t.Run("Given the runtime returns an empty message", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}

		t.Run("Should return a result with a nil message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.Message)
			}
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"boom","message":"runtime down"}`))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t, handler)
			defer WhenCreatingResponseOverHttpAfterEach(suite)

			_, err := suite.sut.Create(context.Background(), request)

			assert.Error(t, err)
		})
	})
}
