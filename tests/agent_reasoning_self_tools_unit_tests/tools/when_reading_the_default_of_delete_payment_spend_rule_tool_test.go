package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfDeletePaymentSpendRuleToolTestingSuite struct {
	sut *tools.DeletePaymentSpendRuleTool
}

func WhenReadingTheDefaultOfDeletePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfDeletePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfDeletePaymentSpendRuleToolTestingSuite{
		sut: tools.NewDeletePaymentSpendRuleTool(
			interfaces_mocks.NewMockDeletePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfDeletePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfDeletePaymentSpendRuleToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
