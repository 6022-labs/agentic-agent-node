package interfaces

import "github.com/ethereum/go-ethereum/common"

// JwtValidatorInterface parses a bearer token and returns the authenticated
// wallet address. Returns an error on signature mismatch, expiry, or malformed
// input, or a missing or revoked session.
type JwtValidatorInterface interface {
	Verify(token string) (common.Address, error)
}
