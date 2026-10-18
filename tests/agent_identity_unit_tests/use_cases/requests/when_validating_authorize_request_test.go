package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingAuthorizeRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank address", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{WalletAddressRequest: requests.WalletAddressRequest{Address: ""}}

		t.Run("Should reject on the address field", func(t *testing.T) {
			t.Parallel()

			assert.ErrorContains(t, req.ValidateAndSanitize(), "address")
		})
	})
}
