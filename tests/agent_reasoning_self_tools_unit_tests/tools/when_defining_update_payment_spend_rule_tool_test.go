package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningUpdatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendRuleTool
}

func WhenDefiningUpdatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenDefiningUpdatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningUpdatePaymentSpendRuleToolTestingSuite{
		sut: tools.NewUpdatePaymentSpendRuleTool(
			interfaces_mocks.NewMockUpdatePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningUpdatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named update_payment_spend_rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningUpdatePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, "update_payment_spend_rule", suite.sut.Definition().Name)
		})

		t.Run("Should require the id, name and config arguments", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningUpdatePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, []string{"id", "name", "config"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
