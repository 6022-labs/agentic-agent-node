package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningCreatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.CreatePaymentSpendRuleTool
}

func WhenDefiningCreatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenDefiningCreatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningCreatePaymentSpendRuleToolTestingSuite{
		sut: tools.NewCreatePaymentSpendRuleTool(
			interfaces_mocks.NewMockCreatePaymentSpendRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningCreatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named create_payment_spend_rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningCreatePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, "create_payment_spend_rule", suite.sut.Definition().Name)
		})

		t.Run("Should require the name and config arguments", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningCreatePaymentSpendRuleToolBeforeEach(t)

			assert.Equal(t, []string{"name", "config"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
