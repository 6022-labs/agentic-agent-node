package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfGetPaymentSpendPolicyToolTestingSuite struct {
	sut *tools.GetPaymentSpendPolicyTool
}

func WhenReadingTheDefaultOfGetPaymentSpendPolicyToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfGetPaymentSpendPolicyToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfGetPaymentSpendPolicyToolTestingSuite{
		sut: tools.NewGetPaymentSpendPolicyTool(
			interfaces_mocks.NewMockGetPaymentSpendPolicyExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfGetPaymentSpendPolicyTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfGetPaymentSpendPolicyToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
