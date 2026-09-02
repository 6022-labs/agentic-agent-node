package x402

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// TransferWithAuthorizationDigest hashes the EIP-3009 typed data under the asset's EIP-712 domain.
func TransferWithAuthorizationDigest(
	auth *Eip3009Authorization,
	asset string,
	eip712Name string,
	eip712Version string,
	chainId int64,
) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"TransferWithAuthorization": {
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              eip712Name,
			Version:           eip712Version,
			ChainId:           math.NewHexOrDecimal256(chainId),
			VerifyingContract: asset,
		},
		Message: apitypes.TypedDataMessage{
			"from":        auth.From,
			"to":          auth.To,
			"value":       auth.Value,
			"validAfter":  auth.ValidAfter,
			"validBefore": auth.ValidBefore,
			"nonce":       auth.Nonce,
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash typed data: %w", err)
	}
	return digest, nil
}

// PermitWitnessTransferFromDigest hashes under Permit2's fixed domain (no version field), witness bound in.
func PermitWitnessTransferFromDigest(auth *Permit2Authorization, chainId int64) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"PermitWitnessTransferFrom": {
				{Name: "permitted", Type: "TokenPermissions"},
				{Name: "spender", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "witness", Type: "Witness"},
			},
			"TokenPermissions": {
				{Name: "token", Type: "address"},
				{Name: "amount", Type: "uint256"},
			},
			"Witness": {
				{Name: "to", Type: "address"},
				{Name: "validAfter", Type: "uint256"},
			},
		},
		PrimaryType: "PermitWitnessTransferFrom",
		Domain: apitypes.TypedDataDomain{
			Name:              Permit2DomainName,
			ChainId:           math.NewHexOrDecimal256(chainId),
			VerifyingContract: Permit2ContractAddress,
		},
		Message: apitypes.TypedDataMessage{
			"permitted": apitypes.TypedDataMessage{
				"token":  auth.Permitted.Token,
				"amount": auth.Permitted.Amount,
			},
			"spender":  auth.Spender,
			"nonce":    auth.Nonce,
			"deadline": auth.Deadline,
			"witness": apitypes.TypedDataMessage{
				"to":         auth.Witness.To,
				"validAfter": auth.Witness.ValidAfter,
			},
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash permit2 typed data: %w", err)
	}
	return digest, nil
}
