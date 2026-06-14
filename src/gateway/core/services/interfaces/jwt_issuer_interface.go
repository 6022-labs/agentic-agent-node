package interfaces

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// JwtIssuerInterface mints HS256 JWTs whose subject is the authenticated
// wallet address.
type JwtIssuerInterface interface {
	Issue(address common.Address) (token string, expiresAt time.Time, err error)
}
