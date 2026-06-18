package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const nonceByteLength = 16

type nonceEntry struct {
	expiresAt time.Time
}

type NonceStore struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.Mutex
	entries map[string]nonceEntry
}

func NewNonceStore(ttl time.Duration) *NonceStore {
	return &NonceStore{
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[string]nonceEntry),
	}
}

func (s *NonceStore) Generate() (string, time.Time, error) {
	bytes := make([]byte, nonceByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}
	nonce := hex.EncodeToString(bytes)
	expiresAt := s.now().Add(s.ttl)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked()
	s.entries[nonce] = nonceEntry{expiresAt: expiresAt}
	return nonce, expiresAt, nil
}

// Consume removes the nonce on success. Returns false if missing or expired.
func (s *NonceStore) Consume(nonce string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[nonce]
	if !ok {
		return false
	}
	delete(s.entries, nonce)
	if s.now().After(entry.expiresAt) {
		return false
	}
	return true
}

func (s *NonceStore) purgeExpiredLocked() {
	now := s.now()
	for k, v := range s.entries {
		if now.After(v.expiresAt) {
			delete(s.entries, k)
		}
	}
}
