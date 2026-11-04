package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfUpdatePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendPolicyTool
}

func WhenReadingTheDefaultOfUpdatePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfUpdatePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfUpdatePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewUpdatePaymentSpendPolicyTool(
			interfaces_mocks.NewMockUpdatePaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfUpdatePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	// Choosing whom it pays is the reason this one is opt-in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfUpdatePaymentSpendPolicyToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
