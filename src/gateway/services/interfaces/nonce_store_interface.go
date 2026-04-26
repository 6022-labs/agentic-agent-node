package interfaces

import "time"

// NonceStoreInterface owns short-lived, single-use challenge nonces. Generate
// returns a fresh nonce + the absolute expiry. Consume looks the nonce up and
// removes it on success — repeated reads return false. Expired nonces are
// treated as missing.
type NonceStoreInterface interface {
	Generate() (nonce string, expiresAt time.Time, err error)
	Consume(nonce string) bool
}
