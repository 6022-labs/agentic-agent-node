package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningSignX402PaymentToolTestingSuite struct {
	sut *tools.SignX402PaymentTool
}

func WhenDefiningSignX402PaymentToolBeforeEach(t *testing.T) *WhenDefiningSignX402PaymentToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningSignX402PaymentToolTestingSuite{
		sut: tools.NewSignX402PaymentTool(interfaces_mocks.NewMockSignX402PaymentExecutorInterface(mockController)),
	}
}

func TestWhenDefiningSignX402PaymentTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named sign_x402_payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningSignX402PaymentToolBeforeEach(t)

			assert.Equal(t, "sign_x402_payment", suite.sut.Definition().Name)
		})

		t.Run("Should require the accept argument", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningSignX402PaymentToolBeforeEach(t)

			assert.Equal(t, []string{"accept"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
