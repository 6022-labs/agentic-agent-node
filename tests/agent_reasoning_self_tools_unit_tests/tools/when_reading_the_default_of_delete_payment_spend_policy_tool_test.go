package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfDeletePaymentSpendPolicyToolTestingSuite struct {
	sut *tools.DeletePaymentSpendPolicyTool
}

func WhenReadingTheDefaultOfDeletePaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfDeletePaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfDeletePaymentSpendPolicyToolTestingSuite{
		sut: tools.NewDeletePaymentSpendPolicyTool(
			interfaces_mocks.NewMockDeletePaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfDeletePaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfDeletePaymentSpendPolicyToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
