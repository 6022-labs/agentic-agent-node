// Package helper holds pure ENS adapter helpers.
package helper

import (
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
)

// ResolveAgentNode computes the agent's ENS namehash, mirroring
// AgentEnsRegistry._resolveAgentNode: parentNode → chainId → collection → agent
// label, each step keccak256(prev || keccak256(labelBytes)).
func ResolveAgentNode(parentNode [32]byte, chainId uint64, collectionName, agentLabel string) [32]byte {
	chainIdLabel := crypto.Keccak256Hash([]byte(strconv.FormatUint(chainId, 10)))
	chainNode := crypto.Keccak256Hash(parentNode[:], chainIdLabel[:])
	collectionLabel := crypto.Keccak256Hash([]byte(collectionName))
	collectionNode := crypto.Keccak256Hash(chainNode[:], collectionLabel[:])
	agentLabelHash := crypto.Keccak256Hash([]byte(agentLabel))
	return crypto.Keccak256Hash(collectionNode[:], agentLabelHash[:])
}
