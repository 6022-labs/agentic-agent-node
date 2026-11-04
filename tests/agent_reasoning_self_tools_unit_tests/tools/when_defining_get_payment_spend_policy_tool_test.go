package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningGetPaymentSpendPolicyToolTestingSuite struct {
	sut *tools.GetPaymentSpendPolicyTool
}

func WhenDefiningGetPaymentSpendPolicyToolBeforeEach(t *testing.T) *WhenDefiningGetPaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningGetPaymentSpendPolicyToolTestingSuite{
		sut: tools.NewGetPaymentSpendPolicyTool(
			interfaces_mocks.NewMockGetPaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningGetPaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named get_payment_spend_policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningGetPaymentSpendPolicyToolBeforeEach(t)

			assert.Equal(t, "get_payment_spend_policy", suite.sut.Definition().Name)
		})
	})
}
