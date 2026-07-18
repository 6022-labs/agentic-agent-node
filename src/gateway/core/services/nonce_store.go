package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

const nonceByteLength = 32

type NonceStore struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.Mutex
	entries map[common.Address]interfaces.ChallengeEntry
}

func NewNonceStore(ttl time.Duration) *NonceStore {
	return &NonceStore{
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[common.Address]interfaces.ChallengeEntry),
	}
}

func (s *NonceStore) Generate(address common.Address, chainId uint64) (*interfaces.ChallengeEntry, error) {
	bytes := make([]byte, nonceByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	now := s.now()
	entry := interfaces.ChallengeEntry{
		Nonce:     hex.EncodeToString(bytes),
		ChainId:   chainId,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.ttl),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked()
	s.entries[address] = entry
	return &entry, nil
}

func (s *NonceStore) Consume(address common.Address) *interfaces.ChallengeEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[address]
	if !ok {
		return nil
	}
	delete(s.entries, address)
	if s.now().After(entry.ExpiresAt) {
		return nil
	}
	return &entry
}

func (s *NonceStore) purgeExpiredLocked() {
	now := s.now()
	for k, v := range s.entries {
		if now.After(v.ExpiresAt) {
			delete(s.entries, k)
		}
	}
}
