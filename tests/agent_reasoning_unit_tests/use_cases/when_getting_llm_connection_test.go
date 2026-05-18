package use_cases_test

import (
	"testing"

	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
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

		t.Run("Should return NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "not_configured", res.Code)
		})
	})

	t.Run("Given the repository returns a connection", func(t *testing.T) {
		t.Parallel()

		stored := model_builders.NewLLMConnectionBuilder().
			WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
			WithModel("gpt-4o").
			Build()
		initSuite := func(suite *WhenGettingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Get().Return(stored, nil)
		}

		t.Run("Should return Ok with the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.LLMConnectionResponse)
			assert.True(t, ok)
			assert.Equal(t, stored.Id, body.Id)
			assert.Equal(t, stored.Model, body.Model)
		})
	})
}
