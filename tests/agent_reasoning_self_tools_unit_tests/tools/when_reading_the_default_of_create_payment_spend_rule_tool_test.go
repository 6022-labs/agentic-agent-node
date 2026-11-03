package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfCreatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.CreatePaymentSpendRuleTool
}

func WhenReadingTheDefaultOfCreatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfCreatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfCreatePaymentSpendRuleToolTestingSuite{
		sut: tools.NewCreatePaymentSpendRuleTool(
			interfaces_mocks.NewMockCreatePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfCreatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	// Enabling this tool lets the agent widen its own limits, so the owner must opt in.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfCreatePaymentSpendRuleToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
