package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingConversationNameTestingSuite struct {
	sut *services.ConversationLabeler

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenGettingConversationNameBeforeEach(t *testing.T) *WhenGettingConversationNameTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	return &WhenGettingConversationNameTestingSuite{
		sut:                        services.NewConversationLabeler(mockConversationRepository, mockKnowledgeMetrics),
		mockConversationRepository: mockConversationRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func TestWhenGettingConversationName(t *testing.T) {
	t.Parallel()

	t.Run("Given loading the conversation fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetName(context.Background(), "c1")

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given the conversation does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "ghost").Return(nil, nil)
		}

		t.Run("Should return no name", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConversationNameBeforeEach(t)
			initSuite(suite)

			name, err := suite.sut.GetName(context.Background(), "ghost")

			assert.NoError(t, err)
			assert.Nil(t, name)
		})
	})

	t.Run("Given a padded conversation id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().
				GetById(gomock.Any(), "c1").
				Return(model_builders.NewConversationBuilder().WithId("c1").Build(), nil)
		}

		t.Run("Should look the conversation up by the trimmed id", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetName(context.Background(), " c1 ")

			assert.NoError(t, err)
		})
	})

	t.Run("Given a named conversation", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").WithName("Audit").Build()

		initSuite := func(suite *WhenGettingConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
		}

		t.Run("Should return its current label", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConversationNameBeforeEach(t)
			initSuite(suite)

			name, err := suite.sut.GetName(context.Background(), "c1")

			assert.NoError(t, err)
			assert.Equal(t, "Audit", *name)
		})
	})
}
