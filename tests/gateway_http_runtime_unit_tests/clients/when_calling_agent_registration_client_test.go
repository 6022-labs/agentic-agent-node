package clients_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/stretchr/testify/assert"
)

type agentRegistrationCaptured struct {
	method string
	path   string
	body   string
}

type WhenCallingAgentRegistrationClientTestingSuite struct {
	sut      *services.HttpAgentRegistrationClient
	server   *httptest.Server
	captured *agentRegistrationCaptured
}

func WhenCallingAgentRegistrationClientBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCallingAgentRegistrationClientTestingSuite {
	t.Helper()
	captured := &agentRegistrationCaptured{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		captured.body = string(body)
		handler(w, r)
	}))
	client, _ := clients.NewClientWithResponses(server.URL)
	return &WhenCallingAgentRegistrationClientTestingSuite{
		sut:      services.NewHttpAgentRegistrationClient(client),
		server:   server,
		captured: captured,
	}
}

func WhenCallingAgentRegistrationClientAfterEach(suite *WhenCallingAgentRegistrationClientTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingAgentRegistrationClient(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts a self-register call", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"mintStatus":"confirmed","ensStatus":"confirmed"}`))
		}

		t.Run("Should POST to /internal/agent/registration/self", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAgentRegistrationClientBeforeEach(t, handler)
			defer WhenCallingAgentRegistrationClientAfterEach(suite)

			payload, err := suite.sut.Self(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, suite.captured.method)
			assert.Equal(t, "/internal/agent/registration/self", suite.captured.path)
			assert.Equal(t, "confirmed", payload.MintStatus)
		})
	})

	t.Run("Given the runtime returns a 409 on ens-info", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"not_minted","message":"mint first"}`))
		}

		t.Run("Should map it to a ConflictError preserving the code", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingAgentRegistrationClientBeforeEach(t, handler)
			defer WhenCallingAgentRegistrationClientAfterEach(suite)

			_, err := suite.sut.EnsInfo(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.True(t, errors.As(err, &conflictErr))
			assert.Equal(t, "not_minted", conflictErr.Code)
		})
	})
}
