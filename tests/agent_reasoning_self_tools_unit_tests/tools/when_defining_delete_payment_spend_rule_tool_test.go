package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningDeletePaymentSpendRuleToolTestingSuite struct {
	sut *tools.DeletePaymentSpendRuleTool
}

func WhenDefiningDeletePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenDefiningDeletePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningDeletePaymentSpendRuleToolTestingSuite{
		sut: tools.NewDeletePaymentSpendRuleTool(
			interfaces_mocks.NewMockDeletePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningDeletePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named delete_payment_spend_rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningDeletePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, "delete_payment_spend_rule", suite.sut.Definition().Name)
		})

		t.Run("Should require the id argument", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningDeletePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, []string{"id"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
