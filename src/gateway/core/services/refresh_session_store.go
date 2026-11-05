package services

import (
	"sync"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
)

type spentRefreshTokenEntry struct {
	spent     interfaces.SpentRefreshToken
	expiresAt time.Time
}

// RefreshSessionStore is in-memory: a restart forgets rotation state, signed sessions stay valid.
type RefreshSessionStore struct {
	now func() time.Time

	mu      sync.Mutex
	spent   map[string]spentRefreshTokenEntry
	revoked map[string]time.Time
}

func NewRefreshSessionStore() *RefreshSessionStore {
	return &RefreshSessionStore{
		now:     time.Now,
		spent:   make(map[string]spentRefreshTokenEntry),
		revoked: make(map[string]time.Time),
	}
}

func (s *RefreshSessionStore) Spend(
	sessionId string,
	tokenId string,
	spent interfaces.SpentRefreshToken,
	expiresAt time.Time,
) (*interfaces.SpentRefreshToken, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked()

	key := sessionId + "/" + tokenId
	if existing, ok := s.spent[key]; ok {
		earlier := existing.spent
		return &earlier, false
	}
	s.spent[key] = spentRefreshTokenEntry{spent: spent, expiresAt: expiresAt}
	return nil, true
}

func (s *RefreshSessionStore) Revoke(sessionId string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked()
	s.revoked[sessionId] = expiresAt
}

func (s *RefreshSessionStore) IsRevoked(sessionId string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt, ok := s.revoked[sessionId]
	return ok && s.now().Before(expiresAt)
}

func (s *RefreshSessionStore) purgeExpiredLocked() {
	now := s.now()
	for k, v := range s.spent {
		if !now.Before(v.expiresAt) {
			delete(s.spent, k)
		}
	}
	for k, v := range s.revoked {
		if !now.Before(v) {
			delete(s.revoked, k)
		}
	}
}
