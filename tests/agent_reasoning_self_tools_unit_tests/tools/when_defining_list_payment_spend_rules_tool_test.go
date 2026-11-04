package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningListPaymentSpendRulesToolTestingSuite struct {
	sut *tools.ListPaymentSpendRulesTool
}

func WhenDefiningListPaymentSpendRulesToolBeforeEach(t *testing.T) *WhenDefiningListPaymentSpendRulesToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningListPaymentSpendRulesToolTestingSuite{
		sut: tools.NewListPaymentSpendRulesTool(
			interfaces_mocks.NewMockListPaymentSpendRulesExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningListPaymentSpendRulesTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named list_payment_spend_rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningListPaymentSpendRulesToolBeforeEach(t)

			assert.Equal(t, "list_payment_spend_rules", suite.sut.Definition().Name)
		})

		t.Run("Should require no argument", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningListPaymentSpendRulesToolBeforeEach(t)

			assert.Empty(t, suite.sut.Definition().InputScheme.Required)
		})
	})
}
