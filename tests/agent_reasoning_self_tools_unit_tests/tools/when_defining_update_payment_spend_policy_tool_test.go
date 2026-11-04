package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningUpdatePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendPolicyTool
}

func WhenDefiningUpdatePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenDefiningUpdatePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningUpdatePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewUpdatePaymentSpendPolicyTool(
			interfaces_mocks.NewMockUpdatePaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningUpdatePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named update_payment_spend_policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningUpdatePaymentSpendPolicyToolBeforeEach(t)

			assert.Equal(t, "update_payment_spend_policy", suite.sut.Definition().Name)
		})

		t.Run("Should require the payToPatterns argument", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningUpdatePaymentSpendPolicyToolBeforeEach(t)

			assert.Equal(t, []string{"payToPatterns"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
