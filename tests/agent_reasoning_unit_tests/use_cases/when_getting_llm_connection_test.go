package use_cases_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingLLMConnectionTestingSuite struct {
	sut *use_cases.GetLLMConnection

	mockRepository *repositories_mocks.MockLLMConnectionRepositoryInterface
}

func WhenGettingLLMConnectionBeforeEach(t *testing.T) *WhenGettingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)

	sut := use_cases.NewGetLLMConnection(mockRepository)

	return &WhenGettingLLMConnectionTestingSuite{
		sut: sut,

		mockRepository: mockRepository,
	}
}

func TestWhenGettingLLMConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the repository returns nil", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Get().Return(nil, nil)
		}

		t.Run("Should return nil without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			got, err := suite.sut.Execute()

			assert.NoError(t, err)
			assert.Nil(t, got)
		})
	})

	t.Run("Given the repository returns a connection", func(t *testing.T) {
		t.Parallel()

		expected := &models.LLMConnection{
			Platform: reasoning_enums.LLMConnectionPlatformOpenAI,
			Model:    "gpt-4o",
		}
		initSuite := func(suite *WhenGettingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Get().Return(expected, nil)
		}

		t.Run("Should pass it through", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			got, err := suite.sut.Execute()

			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	})
}
