package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingTheDefaultOfListPaymentSpendRulesToolTestingSuite struct {
	sut *tools.ListPaymentSpendRulesTool
}

func WhenReadingTheDefaultOfListPaymentSpendRulesToolBeforeEach(
	t *testing.T,
) *WhenReadingTheDefaultOfListPaymentSpendRulesToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenReadingTheDefaultOfListPaymentSpendRulesToolTestingSuite{
		sut: tools.NewListPaymentSpendRulesTool(
			interfaces_mocks.NewMockListPaymentSpendRulesExecutorInterface(mockController),
		),
	}
}

func TestWhenReadingTheDefaultOfListPaymentSpendRulesTool(t *testing.T) {
	t.Parallel()

	// Even reading the policy is opt-in, matching every other payment self tool.
	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be disabled by default", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheDefaultOfListPaymentSpendRulesToolBeforeEach(t)

			assert.False(t, suite.sut.IsEnabledByDefault())
		})
	})
}
