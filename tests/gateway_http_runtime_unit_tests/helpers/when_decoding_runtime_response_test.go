package helpers_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/stretchr/testify/assert"
)

type decodedPayload struct {
	Name string `json:"name"`
}

func TestWhenDecodingRuntimeResponse(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime answers 200 with a JSON body", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode it into out and return no error", func(t *testing.T) {
			t.Parallel()

			var out decodedPayload
			err := helpers.DecodeRuntimeResponse(
				&http.Response{StatusCode: http.StatusOK},
				[]byte(`{"name":"agent"}`),
				&out,
				nil,
			)

			assert.NoError(t, err)
			assert.Equal(t, "agent", out.Name)
		})
	})

	t.Run("Given the runtime answers 204 and no out target is provided", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return no error without decoding", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeRuntimeResponse(&http.Response{StatusCode: http.StatusNoContent}, nil, nil, nil)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the runtime answers 404 and a not-found sentinel is provided", func(t *testing.T) {
		t.Parallel()

		notFound := errors.New("connection not found")

		t.Run("Should return the sentinel error", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeRuntimeResponse(
				&http.Response{StatusCode: http.StatusNotFound},
				[]byte(`{"code":"not_found"}`),
				nil,
				notFound,
			)

			assert.ErrorIs(t, err, notFound)

			var statusErr *helpers.RuntimeStatusError
			assert.False(t, errors.As(err, &statusErr))
		})
	})

	t.Run("Given the runtime answers 404 and no sentinel is provided", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a RuntimeStatusError preserving the status and parsed code", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeRuntimeResponse(
				&http.Response{StatusCode: http.StatusNotFound},
				[]byte(`{"code":"not_found","message":"missing"}`),
				nil,
				nil,
			)

			var statusErr *helpers.RuntimeStatusError
			assert.ErrorAs(t, err, &statusErr)
			assert.Equal(t, http.StatusNotFound, statusErr.StatusCode)
			assert.Equal(t, "not_found", statusErr.Code)
			assert.Equal(t, "missing", statusErr.Message)
		})
	})

	t.Run("Given the runtime answers 400 with a JSON validation body", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a RuntimeStatusError preserving the field and message", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeRuntimeResponse(
				&http.Response{StatusCode: http.StatusBadRequest},
				[]byte(`{"field":"name","message":"required"}`),
				nil,
				nil,
			)

			var statusErr *helpers.RuntimeStatusError
			assert.ErrorAs(t, err, &statusErr)
			assert.Equal(t, http.StatusBadRequest, statusErr.StatusCode)
			assert.Equal(t, "name", statusErr.Field)
			assert.Equal(t, "required", statusErr.Message)
		})
	})

	t.Run("Given the runtime answers 502 with a non-JSON body", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a RuntimeStatusError carrying the raw body and empty fields", func(t *testing.T) {
			t.Parallel()

			err := helpers.DecodeRuntimeResponse(
				&http.Response{StatusCode: http.StatusBadGateway},
				[]byte("upstream exploded"),
				nil,
				nil,
			)

			var statusErr *helpers.RuntimeStatusError
			assert.ErrorAs(t, err, &statusErr)
			assert.Equal(t, http.StatusBadGateway, statusErr.StatusCode)
			assert.Equal(t, "upstream exploded", statusErr.Body)
			assert.Empty(t, statusErr.Code)
			assert.Empty(t, statusErr.Field)
			assert.Empty(t, statusErr.Message)
		})
	})
}
