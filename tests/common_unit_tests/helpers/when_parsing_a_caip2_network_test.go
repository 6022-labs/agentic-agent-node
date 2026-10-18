package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingACaip2Network(t *testing.T) {
	t.Parallel()

	t.Run("Given the eip155 prefix is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			_, err := helpers.ParseCaip2("solana:mainnet")

			assert.ErrorContains(t, err, "CAIP-2 eip155 identifier")
		})
	})

	t.Run("Given the chain id is not a positive integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			for _, in := range []string{"eip155:abc", "eip155:-5", "eip155:0"} {
				_, err := helpers.ParseCaip2(in)
				assert.ErrorContains(t, err, "CAIP-2 eip155 identifier", in)
			}
		})
	})

	t.Run("Given a CAIP-2 eip155 identifier", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the chain id", func(t *testing.T) {
			t.Parallel()

			parsed, err := helpers.ParseCaip2("eip155:80002")

			assert.NoError(t, err)
			assert.Equal(t, uint64(80002), parsed)
		})
	})
}
