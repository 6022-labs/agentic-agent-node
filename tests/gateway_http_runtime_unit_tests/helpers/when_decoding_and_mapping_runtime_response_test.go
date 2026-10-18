package helpers_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenDecodingAndMappingRuntimeResponse(t *testing.T) {
	t.Parallel()

	errNotFound := errors.New("sentinel not found")

	t.Run("Given the runtime answers 404 and a not-found sentinel is provided", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the sentinel untouched", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeAndMapRuntimeResponse(
				&http.Response{StatusCode: http.StatusNotFound},
				[]byte(`{"code":"x"}`),
				nil,
				errNotFound,
			)

			assert.ErrorIs(t, err, errNotFound)
		})
	})

	t.Run("Given the runtime answers 404 and no sentinel is provided", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a not-found error with the runtime code", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeAndMapRuntimeResponse(
				&http.Response{StatusCode: http.StatusNotFound},
				[]byte(`{"code":"mcp_not_found","message":"missing"}`),
				nil,
				nil,
			)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "mcp_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a conflict error", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeAndMapRuntimeResponse(
				&http.Response{StatusCode: http.StatusConflict},
				[]byte(`{"code":"already_exists"}`),
				nil,
				nil,
			)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it to unavailable with runtime_unavailable", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeAndMapRuntimeResponse(
				&http.Response{StatusCode: http.StatusInternalServerError},
				[]byte(`boom`),
				nil,
				nil,
			)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 200 with a JSON body", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode it into the output and return no error", func(t *testing.T) {
			t.Parallel()

			var out struct {
				Name string `json:"name"`
			}

			err := helpers.DecodeAndMapRuntimeResponse(
				&http.Response{StatusCode: http.StatusOK},
				[]byte(`{"name":"agent"}`),
				&out,
				errNotFound,
			)

			assert.NoError(t, err)
			assert.Equal(t, "agent", out.Name)
		})
	})
}
