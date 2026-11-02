package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningUpdatePaymentReceiveRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentReceiveRuleTool
}

func WhenDefiningUpdatePaymentReceiveRuleToolBeforeEach(
	t *testing.T,
) *WhenDefiningUpdatePaymentReceiveRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningUpdatePaymentReceiveRuleToolTestingSuite{
		sut: tools.NewUpdatePaymentReceiveRuleTool(
			interfaces_mocks.NewMockUpdatePaymentReceiveRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningUpdatePaymentReceiveRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should require the id, name, prices and config arguments", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningUpdatePaymentReceiveRuleToolBeforeEach(t)

			assert.Equal(t, []string{"id", "name", "prices", "config"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
