package use_cases_test

import (
	"context"
	"testing"

	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
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
			suite.mockRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "not_configured", notFoundErr.Code)
		})
	})

	t.Run("Given the repository returns a connection", func(t *testing.T) {
		t.Parallel()

		stored := model_builders.NewLLMConnectionBuilder().
			WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
			WithModel("gpt-4o").
			Build()
		initSuite := func(suite *WhenGettingLLMConnectionTestingSuite) {
			suite.mockRepository.EXPECT().Get(gomock.Any()).Return(stored, nil)
		}

		t.Run("Should return Ok with the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, stored.Id, resp.Id)
			assert.Equal(t, stored.Model, resp.Model)
		})
	})
}
