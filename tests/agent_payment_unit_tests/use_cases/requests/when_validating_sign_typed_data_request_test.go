package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSignTypedDataRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given an absent payload", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignTypedDataRequest{}

		t.Run("Should reject on the typedData field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr) {
				assert.Equal(t, "typedData", validationErr.Field)
			}
		})
	})

	t.Run("Given a payload with no primaryType", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignTypedDataRequest{TypedData: &apitypes.TypedData{
			Types: apitypes.Types{"EIP712Domain": {}},
		}}

		t.Run("Should reject on the typedData.primaryType field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr) {
				assert.Equal(t, "typedData.primaryType", validationErr.Field)
			}
		})
	})

	t.Run("Given a primaryType absent from types", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignTypedDataRequest{TypedData: &apitypes.TypedData{
			Types:       apitypes.Types{"EIP712Domain": {}},
			PrimaryType: "Mail",
		}}

		t.Run("Should reject on the typedData.types field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr) {
				assert.Equal(t, "typedData.types", validationErr.Field)
			}
		})
	})

	t.Run("Given a payload with no EIP712Domain", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignTypedDataRequest{TypedData: &apitypes.TypedData{
			Types:       apitypes.Types{"Mail": {}},
			PrimaryType: "Mail",
		}}

		t.Run("Should reject on the typedData.types field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr) {
				assert.Equal(t, "typedData.types", validationErr.Field)
			}
		})
	})

	t.Run("Given a payload declaring its primaryType and EIP712Domain", func(t *testing.T) {
		t.Parallel()
		req := &requests.SignTypedDataRequest{TypedData: &apitypes.TypedData{
			Types:       apitypes.Types{"EIP712Domain": {}, "Mail": {}},
			PrimaryType: "Mail",
		}}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
