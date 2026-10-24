package helpers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	common_helpers "github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

var forwardedCallerId = strings.ToLower(common_tests.RandomEthAddress().Hex())

func TestWhenForwardingCallerId(t *testing.T) {
	t.Parallel()

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should set no caller id header", func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/internal/conversations", nil)

			err := helpers.ForwardCallerId(context.Background(), req)

			assert.NoError(t, err)
			assert.NotContains(t, req.Header, common_helpers.CallerIdHeader)
		})
	})

	t.Run("Given a caller id", func(t *testing.T) {
		t.Parallel()

		ctx := common_helpers.ContextWithCallerId(context.Background(), strings.ToUpper(forwardedCallerId))

		t.Run("Should set it as the caller id header", func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/internal/conversations", nil)

			err := helpers.ForwardCallerId(ctx, req)

			assert.NoError(t, err)
			assert.Equal(t, forwardedCallerId, req.Header.Get(common_helpers.CallerIdHeader))
		})
	})
}
