package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenFormattingACaip2Network(t *testing.T) {
	t.Parallel()

	t.Run("Given a numeric chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should render the CAIP-2 form", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "eip155:80002", helpers.FormatCaip2(80002))
		})
	})
}
