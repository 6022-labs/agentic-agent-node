package interfaces

import "time"

type SpentRefreshToken struct {
	SuccessorId string
	SpentAt     time.Time
}

// RefreshSessionStoreInterface tracks spent refresh tokens and revoked sessions until they expire.
type RefreshSessionStoreInterface interface {
	// Spend atomically marks tokenId spent; if already spent, it returns the earlier record and false.
	Spend(sessionId, tokenId string, spent SpentRefreshToken, expiresAt time.Time) (*SpentRefreshToken, bool)
	Revoke(sessionId string, expiresAt time.Time)
	IsRevoked(sessionId string) bool
}
