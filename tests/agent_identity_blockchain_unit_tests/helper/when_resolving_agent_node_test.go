package helper_test

import (
	"strconv"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestWhenResolvingAgentNode(t *testing.T) {
	t.Parallel()

	t.Run("Given a parent node, chain id, collection and agent label", func(t *testing.T) {
		t.Parallel()

		parentNode := crypto.Keccak256Hash([]byte("agents.eth"))
		chainId := uint64(80002)
		collectionName := "my-collection"
		agentLabel := "my-agent"

		t.Run("Should chain keccak256(prev || keccak256(label)) at each level", func(t *testing.T) {
			t.Parallel()

			chainIdLabel := crypto.Keccak256Hash([]byte(strconv.FormatUint(chainId, 10)))
			chainNode := crypto.Keccak256Hash(parentNode[:], chainIdLabel[:])
			collectionLabel := crypto.Keccak256Hash([]byte(collectionName))
			collectionNode := crypto.Keccak256Hash(chainNode[:], collectionLabel[:])
			agentLabelHash := crypto.Keccak256Hash([]byte(agentLabel))
			expected := crypto.Keccak256Hash(collectionNode[:], agentLabelHash[:])

			node := helper.ResolveAgentNode(parentNode, chainId, collectionName, agentLabel)

			assert.Equal(t, [32]byte(expected), node)
		})

		t.Run("Should be deterministic for identical inputs", func(t *testing.T) {
			t.Parallel()

			first := helper.ResolveAgentNode(parentNode, chainId, collectionName, agentLabel)
			second := helper.ResolveAgentNode(parentNode, chainId, collectionName, agentLabel)

			assert.Equal(t, first, second)
		})

		t.Run("Should change when the agent label changes", func(t *testing.T) {
			t.Parallel()

			node := helper.ResolveAgentNode(parentNode, chainId, collectionName, agentLabel)
			other := helper.ResolveAgentNode(parentNode, chainId, collectionName, "other-agent")

			assert.NotEqual(t, node, other)
		})
	})
}
