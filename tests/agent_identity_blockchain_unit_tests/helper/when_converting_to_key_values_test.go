package helper_test

import (
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/stretchr/testify/assert"
)

func TestWhenConvertingToKeyValues(t *testing.T) {
	t.Parallel()

	t.Run("Given mint key/values", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map each entry preserving key and value", func(t *testing.T) {
			t.Parallel()

			result := helper.ToKeyValues([]identity_interfaces.MintKeyValue{
				{Key: "default", Value: "https://img"},
				{Key: "banner", Value: "https://banner"},
			})

			assert.Len(t, result, 2)
			assert.Equal(t, "default", result[0].Key)
			assert.Equal(t, "https://img", result[0].Value)
			assert.Equal(t, "banner", result[1].Key)
			assert.Equal(t, "https://banner", result[1].Value)
		})
	})

	t.Run("Given no entries", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty, non-nil slice", func(t *testing.T) {
			t.Parallel()

			result := helper.ToKeyValues(nil)

			assert.NotNil(t, result)
			assert.Empty(t, result)
		})
	})
}
