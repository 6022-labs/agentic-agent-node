package interfaces

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type RefreshToken struct {
	Token     string
	ExpiresAt time.Time
}

type RefreshTokenClaims struct {
	Address          common.Address
	SessionId        string
	TokenId          string
	SessionExpiresAt time.Time
}

// RefreshTokenManagerInterface signs refresh tokens with a key access tokens never use.
type RefreshTokenManagerInterface interface {
	Issue(address common.Address) (*RefreshToken, error)
	// Reissue restarts the idle expiry, capped by the session's absolute expiry.
	Reissue(claims *RefreshTokenClaims, tokenId string) (*RefreshToken, error)
	NewTokenId() (string, error)
	Parse(token string) (*RefreshTokenClaims, error)
}
