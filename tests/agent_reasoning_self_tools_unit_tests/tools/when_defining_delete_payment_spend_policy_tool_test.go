package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningDeletePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.DeletePaymentSpendPolicyTool
}

func WhenDefiningDeletePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenDefiningDeletePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningDeletePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewDeletePaymentSpendPolicyTool(
			interfaces_mocks.NewMockDeletePaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningDeletePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named delete_payment_spend_policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningDeletePaymentSpendPolicyToolBeforeEach(t)

			assert.Equal(t, "delete_payment_spend_policy", suite.sut.Definition().Name)
		})
	})
}
