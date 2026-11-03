package tools_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDefiningSignTypedDataToolTestingSuite struct {
	sut *tools.SignTypedDataTool
}

func WhenDefiningSignTypedDataToolBeforeEach(t *testing.T) *WhenDefiningSignTypedDataToolTestingSuite {
	mockController := gomock.NewController(t)

	return &WhenDefiningSignTypedDataToolTestingSuite{
		sut: tools.NewSignTypedDataTool(interfaces_mocks.NewMockSignTypedDataExecutorInterface(mockController)),
	}
}

func TestWhenDefiningSignTypedDataTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the tool is registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be named sign_typed_data", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningSignTypedDataToolBeforeEach(t)

			assert.Equal(t, "sign_typed_data", suite.sut.Definition().Name)
		})

		t.Run("Should require the typedData argument", func(t *testing.T) {
			t.Parallel()

			suite := WhenDefiningSignTypedDataToolBeforeEach(t)

			assert.Equal(t, []string{"typedData"}, suite.sut.Definition().InputScheme.Required)
		})
	})
}
