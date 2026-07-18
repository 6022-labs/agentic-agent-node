package requests_test

import (
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingGetAuthChallengeRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid address with surrounding spaces", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize and pass", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{Address: "  0xd389d997Df5654b87914492E1e0C6861bfe835DE  ", ChainId: 80002}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "0xd389d997Df5654b87914492E1e0C6861bfe835DE", request.Address)
		})
	})

	t.Run("Given a missing chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with a validation error on chainId", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{Address: "0xd389d997Df5654b87914492E1e0C6861bfe835DE"}

			var validationErr *apperrors.ValidationError
			err := request.ValidateAndSanitize()
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	t.Run("Given a malformed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with a validation error on address", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{Address: "0x1234", ChainId: 80002}

			var validationErr *apperrors.ValidationError
			err := request.ValidateAndSanitize()
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "address", validationErr.Field)
		})
	})
}
