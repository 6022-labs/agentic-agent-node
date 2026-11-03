package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfUpdatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendRuleTool
}

func WhenReadingTheDefaultOfUpdatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfUpdatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfUpdatePaymentSpendRuleToolTestingSuite{
		sut: tools.NewUpdatePaymentSpendRuleTool(
			interfaces_mocks.NewMockUpdatePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfUpdatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	// Raising your own cap is the reason this one is opt-in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfUpdatePaymentSpendRuleToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
