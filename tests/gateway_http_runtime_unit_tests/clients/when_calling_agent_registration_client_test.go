package clients_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/stretchr/testify/assert"
)

type agentRegistrationCaptured struct {
	method string
	path   string
	body   string
}

func newAgentRegistrationClient(t *testing.T, handler http.HandlerFunc) (*services.HttpAgentRegistrationClient, *agentRegistrationCaptured, func()) {
	t.Helper()
	captured := &agentRegistrationCaptured{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		captured.body = string(body)
		handler(w, r)
	}))
	httpClient := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{BaseUrl: server.URL, Timeout: 2 * time.Second})
	return services.NewHttpAgentRegistrationClient(httpClient), captured, server.Close
}

func TestWhenCallingAgentRegistrationClient(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts a self-register call", func(t *testing.T) {
		t.Parallel()

		t.Run("Should POST to /internal/agent/registration/self", func(t *testing.T) {
			t.Parallel()

			sut, captured, closeFn := newAgentRegistrationClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"mintStatus":"confirmed","ensStatus":"confirmed"}`))
			})
			defer closeFn()

			payload, err := sut.Self(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, captured.method)
			assert.Equal(t, "/internal/agent/registration/self", captured.path)
			assert.Equal(t, "confirmed", payload.MintStatus)
		})
	})

	t.Run("Given the runtime returns a 409 on ens-info", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a ConflictError preserving the code", func(t *testing.T) {
			t.Parallel()

			sut, _, closeFn := newAgentRegistrationClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"code":"not_minted","message":"mint first"}`))
			})
			defer closeFn()

			_, err := sut.EnsInfo(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.True(t, errors.As(err, &conflictErr))
			assert.Equal(t, "not_minted", conflictErr.Code)
		})
	})
}
