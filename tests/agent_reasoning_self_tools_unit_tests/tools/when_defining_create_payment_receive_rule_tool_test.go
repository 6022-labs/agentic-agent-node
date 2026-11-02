package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningCreatePaymentReceiveRuleToolTestingSuite struct {
	sut *tools.CreatePaymentReceiveRuleTool
}

func WhenDefiningCreatePaymentReceiveRuleToolBeforeEach(
	t *testing.T,
) *WhenDefiningCreatePaymentReceiveRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningCreatePaymentReceiveRuleToolTestingSuite{
		sut: tools.NewCreatePaymentReceiveRuleTool(
			interfaces_mocks.NewMockCreatePaymentReceiveRuleExecutorInterface(mockController),
		),
	}
}

func TestWhenDefiningCreatePaymentReceiveRuleTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should require the name, prices and config arguments", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningCreatePaymentReceiveRuleToolBeforeEach(t)

			assert.Equal(t, []string{"name", "prices", "config"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
