package helper_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// expectedDomainSeparator recomputes the EIP-712 domain separator from the spec, independent of apitypes.
func expectedDomainSeparator(name, version string, chainId int64, asset string) [32]byte {
	typeHash := crypto.Keccak256(
		[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

	var chainWord [32]byte
	big.NewInt(chainId).FillBytes(chainWord[:])

	var addressWord [32]byte
	copy(addressWord[12:], common.HexToAddress(asset).Bytes())

	encoded := make([]byte, 0, 32*5)
	encoded = append(encoded, typeHash...)
	encoded = append(encoded, crypto.Keccak256([]byte(name))...)
	encoded = append(encoded, crypto.Keccak256([]byte(version))...)
	encoded = append(encoded, chainWord[:]...)
	encoded = append(encoded, addressWord[:]...)

	var out [32]byte
	copy(out[:], crypto.Keccak256(encoded))
	return out
}
