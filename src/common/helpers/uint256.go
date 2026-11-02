package helpers

import "math/big"

// maxUint256 is the largest amount an ERC-20 transfer or a varchar(78) column can carry.
var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// ParseUint256 parses a base-10 amount, refusing negatives and anything above 2^256-1.
func ParseUint256(value string) (*big.Int, bool) {
	amount, ok := new(big.Int).SetString(value, 10)
	if !ok || amount.Sign() < 0 || amount.Cmp(maxUint256) > 0 {
		return nil, false
	}
	return amount, true
}
