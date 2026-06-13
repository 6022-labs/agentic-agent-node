package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentRuleRequest := func() *requests.SavePaymentRuleRequest {
		return &requests.SavePaymentRuleRequest{
			Name:            "  Public price  ",
			Network:         "eip155:80002",
			Asset:           "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
			PayTo:           "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			Price:           "10000",
			AddressPatterns: []string{requests.EveryoneAddressPattern, " 0x70997970c51812dc3a010c7d01b50e0d17dc79c8 "},
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize and apply defaults", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "Public price", req.Name)
			assert.Equal(t, ".*", req.AddressPatterns[0])
			assert.Equal(t, "0x70997970c51812dc3a010c7d01b50e0d17dc79c8", req.AddressPatterns[1])
			assert.Equal(t, 60, req.MaxTimeoutSeconds)
			assert.Equal(t, "USDC", req.Eip712Name)
			assert.Equal(t, "2", req.Eip712Version)
		})
	})

	t.Run("Given the rule has no name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Name = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has no address patterns", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.AddressPatterns = nil

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has an invalid regex pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.AddressPatterns = []string{"[invalid"}

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule is free", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept a zero price", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Price = "0"

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has a negative price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Price = "-1"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has a non-eip155 network", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Network = "solana:mainnet"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})
}
