package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// The contract between the two halves of x402: a payment the signer produced must
// recover to the very same wallet. Both sides hash through src/common/x402, so a
// drift here means the shared digest builders moved under one of them.
func TestWhenRecoveringASelfSignedPayer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an eip3009 payment signed by the agent wallet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover as that same wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)

			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "10000",
				AssetTransferMethod: enums.AssetTransferMethodEip3009,
				Eip712Name:          "USDC",
				Eip712Version:       "2",
			})
			assert.NoError(t, err)

			payer, err := services.NewX402PaymentVerifier().RecoverPayer(
				suite.decodePaymentPayload(t, signed.PaymentSignature),
				&models.PaymentReceiveRuleX402Config{
					Network:             signingNetwork,
					Asset:               signingAsset,
					PayTo:               signingPayTo,
					AssetTransferMethod: enums.AssetTransferMethodEip3009,
					Eip712Name:          lo.ToPtr("USDC"),
					Eip712Version:       lo.ToPtr("2"),
				},
			)

			assert.NoError(t, err)
			assert.Equal(t, suite.walletAddress, payer)
		})
	})

	t.Run("Given a permit2 payment signed by the agent wallet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover as that same wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningX402PaymentBeforeEach(t)

			// Zero-price keeps the on-chain approval out of a unit test.
			signed, err := suite.sut.Sign(ctx, &payment_interfaces.X402PaymentQuote{
				Network:             signingNetwork,
				Asset:               signingAsset,
				PayTo:               signingPayTo,
				Amount:              "0",
				AssetTransferMethod: enums.AssetTransferMethodPermit2,
			})
			assert.NoError(t, err)

			payer, err := services.NewX402PaymentVerifier().RecoverPayer(
				suite.decodePaymentPayload(t, signed.PaymentSignature),
				&models.PaymentReceiveRuleX402Config{
					Network:             signingNetwork,
					Asset:               signingAsset,
					PayTo:               signingPayTo,
					AssetTransferMethod: enums.AssetTransferMethodPermit2,
				},
			)

			assert.NoError(t, err)
			assert.Equal(t, suite.walletAddress, payer)
		})
	})
}
