package interfaces

import "time"

// NonceStoreInterface owns short-lived, single-use challenge nonces. Consume
// removes on first success; repeated or expired reads return false.
type NonceStoreInterface interface {
	Generate() (nonce string, expiresAt time.Time, err error)
	Consume(nonce string) bool
}
