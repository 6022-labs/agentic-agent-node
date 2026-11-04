package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfSignX402PaymentToolTestingSuite struct {
	sut *tools.SignX402PaymentTool
}

func WhenReadingTheDefaultOfSignX402PaymentToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfSignX402PaymentToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfSignX402PaymentToolTestingSuite{
		sut: tools.NewSignX402PaymentTool(
			interfaces_mocks.NewMockSignX402PaymentExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfSignX402PaymentTool(t *testing.T) {
	t.Parallel()

	// It moves the wallet's funds, so the owner must opt in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfSignX402PaymentToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
