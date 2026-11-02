package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_tests "github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSignX402PaymentRequest(t *testing.T) {
	t.Parallel()

	generateSignX402PaymentRequest := func() *requests.SignX402PaymentRequest {
		return &requests.SignX402PaymentRequest{
			Network:             "eip155:8453",
			Asset:               common_tests.RandomEthAddress().Hex(),
			PayTo:               common_tests.RandomEthAddress().Hex(),
			Amount:              "10000",
			AssetTransferMethod: "permit2",
		}
	}

	t.Run("Given a network that is not CAIP-2", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the network field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Network = "base"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "network", validationErr.Field)
		})
	})

	// A spend rule is keyed on the canonical network, so a padded chain id must not slip past it.
	t.Run("Given a network whose chain id has leading zeros", func(t *testing.T) {
		t.Parallel()

		t.Run("Should canonicalize the network", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Network = " eip155:08453 "

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "eip155:8453", request.Network)
		})
	})

	t.Run("Given an asset that is not an address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the asset field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Asset = "0x1234"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "asset", validationErr.Field)
		})
	})

	t.Run("Given a payee that is not an address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the payTo field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.PayTo = "0x1234"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "payTo", validationErr.Field)
		})
	})

	t.Run("Given an amount that is not an integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the amount field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Amount = "1.5"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "amount", validationErr.Field)
		})
	})

	t.Run("Given a negative amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the amount field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Amount = "-1"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "amount", validationErr.Field)
		})
	})

	t.Run("Given an amount above the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the amount field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.Amount = "115792089237316195423570985008687907853269984665640564039457584007913129639936"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "amount", validationErr.Field)
		})
	})

	t.Run("Given a quote carrying its method and domain under extra", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read them from extra", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.AssetTransferMethod = ""
			request.Extra = map[string]any{"assetTransferMethod": "eip3009", "name": "USD Coin", "version": "2"}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "eip3009", request.AssetTransferMethod)
			assert.Equal(t, "USD Coin", request.Eip712Name)
			assert.Equal(t, "2", request.Eip712Version)
		})
	})

	t.Run("Given a quote that names no transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default to eip3009", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.AssetTransferMethod = ""
			request.Eip712Name = "USD Coin"
			request.Eip712Version = "2"

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "eip3009", request.AssetTransferMethod)
		})
	})

	t.Run("Given an unsupported transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the assetTransferMethod field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.AssetTransferMethod = "permit"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "assetTransferMethod", validationErr.Field)
		})
	})

	t.Run("Given an eip3009 quote without its EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the eip712Name field", func(t *testing.T) {
			t.Parallel()

			request := generateSignX402PaymentRequest()
			request.AssetTransferMethod = "eip3009"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "eip712Name", validationErr.Field)
		})
	})
}
