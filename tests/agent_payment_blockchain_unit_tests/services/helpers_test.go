package services_test

import (
	"math/big"
	"strings"
	"testing"

	contract_abi "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// packErc20Output encodes a method return the way the contract would.
func packErc20Output(t *testing.T, method string, values ...any) []byte {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(contract_abi.Erc20MetaData.ABI))
	assert.NoError(t, err)
	packed, err := parsed.Methods[method].Outputs.Pack(values...)
	assert.NoError(t, err)
	return packed
}

// erc20Probes are the three decode-checked ERC-20 reads the validator runs
// before resolving the EIP-712 domain.
func erc20Probes(t *testing.T, suite *WhenValidatingErc20TestingSuite) []any {
	t.Helper()
	return []any{
		suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(packErc20Output(t, "decimals", uint8(6)), nil),
		suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(packErc20Output(t, "symbol", "USDC"), nil),
		suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(packErc20Output(t, "totalSupply", big.NewInt(1_000_000)), nil),
	}
}

// eip712DomainSeparator hashes the domain the way the token does — repeated
// here so a name()/version() fallback can be verified against it.
func eip712DomainSeparator(t *testing.T, name, version string) [32]byte {
	t.Helper()
	domain := apitypes.TypedDataDomain{
		Name:              name,
		Version:           version,
		ChainId:           math.NewHexOrDecimal256(erc20TestChainId),
		VerifyingContract: erc20TestAsset,
	}
	typedData := apitypes.TypedData{
		Types: apitypes.Types{"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		}},
		Domain: domain,
	}
	hash, err := typedData.HashStruct("EIP712Domain", domain.Map())
	assert.NoError(t, err)
	var out [32]byte
	copy(out[:], hash)
	return out
}
