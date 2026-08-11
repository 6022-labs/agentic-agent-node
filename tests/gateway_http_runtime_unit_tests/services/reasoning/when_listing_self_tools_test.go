package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingSelfToolsTestingSuite struct {
	sut    *services.HttpSelfToolConfigClient
	server *httptest.Server

	handler http.HandlerFunc
}

func WhenListingSelfToolsBeforeEach(t *testing.T) *WhenListingSelfToolsTestingSuite {
	suite := &WhenListingSelfToolsTestingSuite{}

	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if suite.handler == nil {
			http.NotFound(w, r)
			return
		}
		suite.handler(w, r)
	}))

	client, err := clients.NewClientWithResponses(suite.server.URL)
	require.NoError(t, err)

	suite.sut = services.NewHttpSelfToolConfigClient(client)

	return suite
}

func WhenListingSelfToolsAfterEach(suite *WhenListingSelfToolsTestingSuite) {
	suite.server.Close()
}

func TestWhenListingSelfTools(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the configured tools", func(t *testing.T) {
		t.Parallel()

		var capturedMethod string
		var capturedPath string

		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.handler = func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"tools":[{"id":"tool-1","name":"run_command","description":"runs a command","enabled":true}]}`))
			}
		}

		t.Run("Should GET the list and decode the payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			defer WhenListingSelfToolsAfterEach(suite)
			initSuite(suite)

			payload, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, http.MethodGet, capturedMethod)
			assert.Equal(t, "/internal/config/self-tools", capturedPath)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Tools, 1) {
				assert.Equal(t, "tool-1", payload.Tools[0].Id)
				assert.Equal(t, "run_command", payload.Tools[0].Name)
				assert.True(t, payload.Tools[0].Enabled)
			}
		})
	})

	t.Run("Given the runtime returns non-2xx", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.handler = func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			defer WhenListingSelfToolsAfterEach(suite)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			assert.ErrorContains(t, err, "500")
		})
	})
}
