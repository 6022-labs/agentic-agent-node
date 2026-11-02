package services_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

const quotedNetwork = "eip155:8453"

type WhenQuotingTypedDataSpendTestingSuite struct {
	sut *services.TypedDataSpendQuoter
}

func WhenQuotingTypedDataSpendBeforeEach() *WhenQuotingTypedDataSpendTestingSuite {
	return &WhenQuotingTypedDataSpendTestingSuite{sut: services.NewTypedDataSpendQuoter()}
}

func spendTypedData(
	primaryType string,
	verifyingContract string,
	message apitypes.TypedDataMessage,
) *apitypes.TypedData {
	return &apitypes.TypedData{
		PrimaryType: primaryType,
		Domain: apitypes.TypedDataDomain{
			ChainId:           math.NewHexOrDecimal256(8453),
			VerifyingContract: verifyingContract,
		},
		Message: message,
	}
}

func TestWhenQuotingTypedDataSpend(t *testing.T) {
	t.Parallel()

	token := common_tests.RandomEthAddress().Hex()
	payee := common_tests.RandomEthAddress().Hex()

	t.Run("Given a TransferWithAuthorization payload", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})

		t.Run("Should quote the token, the recipient and the value as an eip3009 payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, quotedNetwork, quote.Network)
				assert.Equal(t, token, quote.Asset)
				assert.Equal(t, payee, quote.PayTo)
				assert.Equal(t, "10000", quote.Amount)
				assert.Equal(t, enums.AssetTransferMethodEip3009, quote.AssetTransferMethod)
			}
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a ReceiveWithAuthorization payload", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("ReceiveWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})

		t.Run("Should quote it as an eip3009 payment to the recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, payee, quote.PayTo)
				assert.Equal(t, enums.AssetTransferMethodEip3009, quote.AssetTransferMethod)
			}
		})
	})

	t.Run("Given an EIP-2612 permit", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"value":   "5000",
		})

		t.Run("Should quote the allowance to the spender with no transfer method", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, token, quote.Asset)
				assert.Equal(t, payee, quote.PayTo)
				assert.Equal(t, enums.AssetTransferMethod(""), quote.AssetTransferMethod)
			}
			assert.Equal(t, big.NewInt(5000), amount)
		})
	})

	t.Run("Given a DAI-style permit that grants an allowance", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"allowed": true,
		})

		t.Run("Should quote the largest uint256", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, math.MaxBig256, amount)
		})
	})

	t.Run("Given a DAI-style permit that revokes an allowance", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"allowed": false,
		})

		t.Run("Should quote nothing to spend", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(0), amount)
		})
	})

	t.Run("Given a permit2 witness transfer through the x402 proxy", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitWitnessTransferFrom", x402.Permit2ContractAddress, apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{"token": token, "amount": "7000"},
			"spender":   x402.Permit2Spender,
			"witness":   apitypes.TypedDataMessage{"to": payee},
		})

		t.Run("Should quote the witness recipient as a permit2 payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, token, quote.Asset)
				assert.Equal(t, payee, quote.PayTo)
				assert.Equal(t, enums.AssetTransferMethodPermit2, quote.AssetTransferMethod)
			}
			assert.Equal(t, big.NewInt(7000), amount)
		})
	})

	t.Run("Given a permit2 transfer to another spender", func(t *testing.T) {
		t.Parallel()

		spender := common_tests.RandomEthAddress().Hex()
		typedData := spendTypedData("PermitTransferFrom", x402.Permit2ContractAddress, apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{"token": token, "amount": "7000"},
			"spender":   spender,
		})

		// Outside the x402 proxy the spender picks the recipient, so the spender is who gets paid.
		t.Run("Should quote the spender as the payee", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, spender, quote.PayTo)
			}
		})
	})

	t.Run("Given a permit2 allowance for one token", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitSingle", x402.Permit2ContractAddress, apitypes.TypedDataMessage{
			"details": apitypes.TypedDataMessage{"token": token, "amount": "8000"},
			"spender": payee,
		})

		t.Run("Should quote the allowance to the spender as a permit2 payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			if assert.NotNil(t, quote) {
				assert.Equal(t, token, quote.Asset)
				assert.Equal(t, payee, quote.PayTo)
				assert.Equal(t, enums.AssetTransferMethodPermit2, quote.AssetTransferMethod)
			}
			assert.Equal(t, big.NewInt(8000), amount)
		})
	})

	t.Run("Given a batch permit", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitBatchTransferFrom", x402.Permit2ContractAddress, apitypes.TypedDataMessage{})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.Nil(t, quote)
			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.primaryType", validationError.Field)
			}
		})
	})

	t.Run("Given a payload that moves no funds", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Mail", "", apitypes.TypedDataMessage{"contents": "hello"})

		t.Run("Should return no quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Nil(t, quote)
			assert.Nil(t, amount)
		})
	})

	t.Run("Given a token payload without a chain id", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Domain.ChainId = nil

		t.Run("Should refuse it on the chain id", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain.chainId", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload on chain zero", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Domain.ChainId = math.NewHexOrDecimal256(0)

		t.Run("Should refuse it on the chain id", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain.chainId", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload with a malformed token address", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", "not-an-address", apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload without a recipient", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{"value": "10000"})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload without a value", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{"to": payee})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is above uint256", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": new(big.Int).Add(math.MaxBig256, big.NewInt(1)).String(),
		})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is negative", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "-1",
		})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is not a number", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "ten",
		})

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is hex", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "0x2710",
		})

		t.Run("Should read the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a token payload whose value is a json number", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": json.Number("10000"),
		})

		t.Run("Should read the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a token payload whose value is a whole float", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": float64(10000),
		})

		t.Run("Should read the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a token payload whose value is a float beyond 2^53", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": float64(1 << 60),
		})

		// A float64 that large no longer holds the integer the caller meant.
		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is a big integer", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": big.NewInt(10000),
		})

		t.Run("Should read the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a token payload whose value is a hex-or-decimal integer", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": math.NewHexOrDecimal256(10000),
		})

		t.Run("Should read the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})
}
