package interfaces

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// JwtIssuerInterface mints HS256 JWTs whose subject is the authenticated
// wallet address and whose sid ties them to a refresh session they never outlive.
type JwtIssuerInterface interface {
	Issue(
		address common.Address,
		sessionId string,
		sessionExpiresAt time.Time,
	) (token string, expiresAt time.Time, err error)
}
