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
	types := standardTokenTypes()
	types["EIP712Domain"] = tokenDomainTypes()
	return &apitypes.TypedData{
		Types:       types,
		PrimaryType: primaryType,
		Domain: apitypes.TypedDataDomain{
			ChainId:           math.NewHexOrDecimal256(8453),
			VerifyingContract: verifyingContract,
		},
		Message: message,
	}
}

// permit2SpendTypedData signs over Permit2's own domain, as a real Permit2 payload does.
func permit2SpendTypedData(primaryType string, message apitypes.TypedDataMessage) *apitypes.TypedData {
	typedData := spendTypedData(primaryType, x402.Permit2ContractAddress, message)
	typedData.Domain.Name = x402.Permit2DomainName
	return typedData
}

// standardTokenTypes declares the gated token payloads the way their contracts hash them.
func standardTokenTypes() apitypes.Types {
	authorization := []apitypes.Type{
		{Name: "from", Type: "address"},
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "validAfter", Type: "uint256"},
		{Name: "validBefore", Type: "uint256"},
		{Name: "nonce", Type: "bytes32"},
	}
	return apitypes.Types{
		"TransferWithAuthorization": authorization,
		"ReceiveWithAuthorization":  authorization,
		"TokenPermissions":          {{Name: "token", Type: "address"}, {Name: "amount", Type: "uint256"}},
		"Witness":                   {{Name: "to", Type: "address"}, {Name: "validAfter", Type: "uint256"}},
		"PermitTransferFrom": {
			{Name: "permitted", Type: "TokenPermissions"},
			{Name: "spender", Type: "address"},
			{Name: "nonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
		"PermitWitnessTransferFrom": {
			{Name: "permitted", Type: "TokenPermissions"},
			{Name: "spender", Type: "address"},
			{Name: "nonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
			{Name: "witness", Type: "Witness"},
		},
		"PermitDetails": {
			{Name: "token", Type: "address"},
			{Name: "amount", Type: "uint160"},
			{Name: "expiration", Type: "uint48"},
			{Name: "nonce", Type: "uint48"},
		},
		"PermitSingle": {
			{Name: "details", Type: "PermitDetails"},
			{Name: "spender", Type: "address"},
			{Name: "sigDeadline", Type: "uint256"},
		},
	}
}

// tokenDomainTypes declares the domain fields a token contract hashes.
func tokenDomainTypes() []apitypes.Type {
	return []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}
}

// eip2612PermitTypes declares the standard EIP-2612 permit.
func eip2612PermitTypes() apitypes.Types {
	return apitypes.Types{
		"EIP712Domain": tokenDomainTypes(),
		"Permit": {
			{Name: "owner", Type: "address"},
			{Name: "spender", Type: "address"},
			{Name: "value", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
	}
}

// daiPermitTypes declares DAI's permit, whose allowed flag replaces the EIP-2612 value.
func daiPermitTypes() apitypes.Types {
	return apitypes.Types{
		"EIP712Domain": tokenDomainTypes(),
		"Permit": {
			{Name: "holder", Type: "address"},
			{Name: "spender", Type: "address"},
			{Name: "nonce", Type: "uint256"},
			{Name: "expiry", Type: "uint256"},
			{Name: "allowed", Type: "bool"},
		},
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
		typedData.Types = eip2612PermitTypes()

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
		typedData.Types = daiPermitTypes()

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
		typedData.Types = daiPermitTypes()

		t.Run("Should quote nothing to spend", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(0), amount)
		})
	})

	// A revoke flag must not hide a value the same permit also grants.
	t.Run("Given a DAI-style permit that also declares a value and revokes", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"value":   "7000",
			"allowed": false,
		})
		typedData.Types = apitypes.Types{
			"EIP712Domain": tokenDomainTypes(),
			"Permit": {
				{Name: "holder", Type: "address"},
				{Name: "spender", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "allowed", Type: "bool"},
			},
		}

		t.Run("Should quote the declared value", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(7000), amount)
		})
	})

	// The signed type, not the message, says whether the permit is DAI-style.
	t.Run("Given an EIP-2612 permit whose message adds an undeclared allowed flag", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"value":   "5000",
			"allowed": false,
		})
		typedData.Types = eip2612PermitTypes()

		t.Run("Should quote its value", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(5000), amount)
		})
	})

	t.Run("Given a permit that declares allowed as an amount and no value", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("Permit", token, apitypes.TypedDataMessage{
			"spender": payee,
			"allowed": "1000",
		})
		typedData.Types = apitypes.Types{
			"EIP712Domain": tokenDomainTypes(),
			"Permit": {
				{Name: "owner", Type: "address"},
				{Name: "spender", Type: "address"},
				{Name: "allowed", Type: "uint256"},
			},
		}

		t.Run("Should refuse the permit with no readable amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.Nil(t, quote)
			var validationErr *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t, "typedData", validationErr.Field)
			}
		})
	})

	t.Run("Given a permit2 witness transfer through the x402 proxy", func(t *testing.T) {
		t.Parallel()

		typedData := permit2SpendTypedData("PermitWitnessTransferFrom", apitypes.TypedDataMessage{
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
		typedData := permit2SpendTypedData("PermitTransferFrom", apitypes.TypedDataMessage{
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

		typedData := permit2SpendTypedData("PermitSingle", apitypes.TypedDataMessage{
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

	t.Run("Given a permit2 transfer signed for another contract", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData(
			"PermitWitnessTransferFrom",
			common_tests.RandomEthAddress().Hex(),
			apitypes.TypedDataMessage{
				"permitted": apitypes.TypedDataMessage{"token": token, "amount": "0"},
				"spender":   x402.Permit2Spender,
				"witness":   apitypes.TypedDataMessage{"to": payee},
			},
		)
		typedData.Domain.Name = x402.Permit2DomainName

		t.Run("Should refuse it on the domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.Nil(t, quote)
			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain", validationError.Field)
			}
		})
	})

	t.Run("Given a permit2 allowance signed for another contract", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitSingle", common_tests.RandomEthAddress().Hex(), apitypes.TypedDataMessage{
			"details": apitypes.TypedDataMessage{"token": token, "amount": "0"},
			"spender": payee,
		})
		typedData.Domain.Name = x402.Permit2DomainName

		t.Run("Should refuse it on the domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.Nil(t, quote)
			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain", validationError.Field)
			}
		})
	})

	t.Run("Given a permit2 transfer under another domain name", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitTransferFrom", x402.Permit2ContractAddress, apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{"token": token, "amount": "7000"},
			"spender":   payee,
		})
		typedData.Domain.Name = "NotPermit2"

		t.Run("Should refuse it on the domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			quote, _, err := suite.sut.Quote(typedData)

			assert.Nil(t, quote)
			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain", validationError.Field)
			}
		})
	})

	t.Run("Given a batch permit", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("PermitBatchTransferFrom", x402.Permit2ContractAddress, apitypes.TypedDataMessage{})

		t.Run("Should refuse the batch permit", func(t *testing.T) {
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

	t.Run("Given a token payload on a chain id beyond 64 bits", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Domain.ChainId = (*math.HexOrDecimal256)(new(big.Int).Lsh(big.NewInt(1), 64))

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

	t.Run("Given a token payload whose domain does not declare its contract", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Types["EIP712Domain"] = []apitypes.Type{
			{Name: "name", Type: "string"},
			{Name: "chainId", Type: "uint256"},
		}

		t.Run("Should refuse it on the domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose domain declares its chain id twice", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Types["EIP712Domain"] = []apitypes.Type{
			{Name: "chainId", Type: "uint256"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		}

		t.Run("Should refuse it on the domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData.domain", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload with a malformed token address", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", "not-an-address", apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})

		t.Run("Should refuse the malformed token address", func(t *testing.T) {
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

		t.Run("Should refuse the payload without a recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose types do not declare its recipient", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": "10000",
		})
		typedData.Types["TransferWithAuthorization"] = []apitypes.Type{{Name: "value", Type: "uint256"}}

		t.Run("Should refuse the undeclared recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a permit2 witness transfer whose types do not declare the witness", func(t *testing.T) {
		t.Parallel()

		typedData := permit2SpendTypedData("PermitWitnessTransferFrom", apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{"token": token, "amount": "7000"},
			"spender":   x402.Permit2Spender,
			"witness":   apitypes.TypedDataMessage{"to": payee},
		})
		typedData.Types["PermitWitnessTransferFrom"] = typedData.Types["PermitTransferFrom"]

		t.Run("Should refuse the undeclared witness recipient", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload with a malformed recipient", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    "not-an-address",
			"value": "10000",
		})

		t.Run("Should refuse the malformed recipient", func(t *testing.T) {
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

		t.Run("Should refuse the payload without a value", func(t *testing.T) {
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

		t.Run("Should refuse the value above uint256", func(t *testing.T) {
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

		t.Run("Should refuse the negative value", func(t *testing.T) {
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

		t.Run("Should refuse the value that is not a number", func(t *testing.T) {
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

		t.Run("Should read the hex value", func(t *testing.T) {
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

		t.Run("Should read the json number", func(t *testing.T) {
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

		t.Run("Should read the whole float", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})

	t.Run("Given a token payload whose value is a fractional float", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": 1.5,
		})

		t.Run("Should refuse the fractional amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, _, err := suite.sut.Quote(typedData)

			var validationError *apperrors.ValidationError
			if assert.ErrorAs(t, err, &validationError) {
				assert.Equal(t, "typedData", validationError.Field)
			}
		})
	})

	t.Run("Given a token payload whose value is a float beyond 2^53", func(t *testing.T) {
		t.Parallel()

		typedData := spendTypedData("TransferWithAuthorization", token, apitypes.TypedDataMessage{
			"to":    payee,
			"value": float64(1 << 60),
		})

		// A float64 that large no longer holds the integer the caller meant.
		t.Run("Should refuse the float beyond 2^53", func(t *testing.T) {
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

		t.Run("Should read the big integer", func(t *testing.T) {
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

		t.Run("Should read the hex-or-decimal integer", func(t *testing.T) {
			t.Parallel()

			suite := WhenQuotingTypedDataSpendBeforeEach()

			_, amount, err := suite.sut.Quote(typedData)

			assert.NoError(t, err)
			assert.Equal(t, big.NewInt(10000), amount)
		})
	})
}
