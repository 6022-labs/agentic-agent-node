package clients_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenFetchingAgentCardTestingSuite struct {
	sut    *clients.A2AClient
	server *httptest.Server

	capturedMethod string
	capturedPath   string
}

func WhenFetchingAgentCardBeforeEach(
	t *testing.T,
	cardTimeoutSeconds int,
	respond http.HandlerFunc,
) *WhenFetchingAgentCardTestingSuite {
	t.Helper()

	suite := &WhenFetchingAgentCardTestingSuite{
		sut: clients.NewA2AClient(clients.NewA2AClientParams{
			Logger: zap.NewNop(),
			A2ASettings: &settings.A2ASettings{
				RequestTimeoutSeconds: 5,
				CardTimeoutSeconds:    cardTimeoutSeconds,
				MaxResponseBytes:      1 << 20,
			},
			HTTPClient: http.DefaultClient,
		}),
	}
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.capturedMethod = r.Method
		suite.capturedPath = r.URL.Path
		respond(w, r)
	}))
	return suite
}

func WhenFetchingAgentCardAfterEach(suite *WhenFetchingAgentCardTestingSuite) {
	suite.server.Close()
}

func TestWhenFetchingAgentCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the agent never answers", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Second):
			}
		}

		t.Run("Should give up after the card timeout", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t, 1, respond)
			defer WhenFetchingAgentCardAfterEach(suite)

			_, err := suite.sut.FetchAgentCard(ctx, suite.server.URL)

			assert.ErrorIs(t, err, context.DeadlineExceeded)
		})
	})

	t.Run("Given the agent serves a card", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"hermes","supportedInterfaces":[` +
				`{"url":"https://peer/grpc","protocolBinding":"GRPC"},` +
				`{"url":"https://peer/api/a2a","protocolBinding":"JSONRPC","protocolVersion":"1.0"}]}`))
		}

		t.Run("Should read it from the well-known path", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t, 5, respond)
			defer WhenFetchingAgentCardAfterEach(suite)

			_, err := suite.sut.FetchAgentCard(ctx, suite.server.URL+"/")

			assert.NoError(t, err)
			assert.Equal(t, http.MethodGet, suite.capturedMethod)
			assert.Equal(t, "/.well-known/agent-card.json", suite.capturedPath)
		})

		t.Run("Should decode the card", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t, 5, respond)
			defer WhenFetchingAgentCardAfterEach(suite)

			card, err := suite.sut.FetchAgentCard(ctx, suite.server.URL)

			assert.NoError(t, err)
			assert.Equal(t, &models.ExternalAgentCard{
				Name: "hermes",
				SupportedInterfaces: []models.ExternalAgentInterface{
					{Url: "https://peer/grpc", ProtocolBinding: "GRPC"},
					{Url: "https://peer/api/a2a", ProtocolBinding: "JSONRPC", ProtocolVersion: "1.0"},
				},
			}, card)
		})
	})

	t.Run("Given the agent serves no card", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`not found`))
		}

		t.Run("Should report the status", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t, 5, respond)
			defer WhenFetchingAgentCardAfterEach(suite)

			_, err := suite.sut.FetchAgentCard(ctx, suite.server.URL)

			assert.ErrorContains(t, err, "agent card request returned 404")
		})
	})

	t.Run("Given the agent serves an undecodable card", func(t *testing.T) {
		t.Parallel()

		respond := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`not json`))
		}

		t.Run("Should report the decode failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t, 5, respond)
			defer WhenFetchingAgentCardAfterEach(suite)

			_, err := suite.sut.FetchAgentCard(ctx, suite.server.URL)

			assert.ErrorContains(t, err, "failed to decode agent card")
		})
	})
}
