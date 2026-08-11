package services_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenUpdatingSelfToolTestingSuite struct {
	sut    *services.HttpSelfToolConfigClient
	server *httptest.Server

	handler http.HandlerFunc
}

func WhenUpdatingSelfToolBeforeEach(t *testing.T) *WhenUpdatingSelfToolTestingSuite {
	suite := &WhenUpdatingSelfToolTestingSuite{}

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

func WhenUpdatingSelfToolAfterEach(suite *WhenUpdatingSelfToolTestingSuite) {
	suite.server.Close()
}

func TestWhenUpdatingSelfTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		var capturedMethod string
		var capturedPath string
		var capturedContentType string
		var capturedBody []byte

		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.handler = func(w http.ResponseWriter, r *http.Request) {
				capturedMethod = r.Method
				capturedPath = r.URL.Path
				capturedContentType = r.Header.Get("Content-Type")
				capturedBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"` + id.String() + `","name":"run_command","enabled":false}`))
			}
		}

		t.Run("Should PUT the flag and decode the updated tool", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			defer WhenUpdatingSelfToolAfterEach(suite)
			initSuite(suite)

			payload, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateSelfToolPayload{Enabled: false})

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPut, capturedMethod)
			assert.Equal(t, "/internal/config/self-tools/"+id.String(), capturedPath)
			assert.Equal(t, "application/json", capturedContentType)
			assert.JSONEq(t, `{"enabled":false}`, string(capturedBody))
			if assert.NotNil(t, payload) {
				assert.Equal(t, id.String(), payload.Id)
				assert.False(t, payload.Enabled)
			}
		})
	})

	t.Run("Given updating a missing tool", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.handler = func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":"not_found"}`))
			}
		}

		t.Run("Should return ErrSelfToolNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			defer WhenUpdatingSelfToolAfterEach(suite)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), uuid.New(), &models.ExternalUpdateSelfToolPayload{Enabled: true})

			assert.ErrorIs(t, err, gateway_interfaces.ErrSelfToolNotFound)
		})
	})

	t.Run("Given the runtime rejects the body", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.handler = func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"field":"enabled","message":"enabled is required"}`))
			}
		}

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			defer WhenUpdatingSelfToolAfterEach(suite)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), uuid.New(), &models.ExternalUpdateSelfToolPayload{Enabled: true})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "enabled", validationErr.Field)
		})
	})
}
