// Package helper holds pure agent_payment_blockchain adapter helpers.
package helper

import (
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ComputeDomainSeparator hashes the EIP-712 domain the way the token does, to
// verify a name()/version() guess against the on-chain DOMAIN_SEPARATOR().
func ComputeDomainSeparator(name, version string, chainId int64, asset string) [32]byte {
	domain := apitypes.TypedDataDomain{
		Name:              name,
		Version:           version,
		ChainId:           math.NewHexOrDecimal256(chainId),
		VerifyingContract: asset,
	}
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		Domain: domain,
	}
	hash, err := typedData.HashStruct("EIP712Domain", domain.Map())
	if err != nil {
		return [32]byte{}
	}
	var out [32]byte
	copy(out[:], hash)
	return out
}
