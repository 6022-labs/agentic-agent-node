package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenLabellingConversationTestingSuite struct {
	sut *services.ConversationLabeler

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenLabellingConversationBeforeEach(t *testing.T) *WhenLabellingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	return &WhenLabellingConversationTestingSuite{
		sut:                        services.NewConversationLabeler(mockConversationRepository, mockKnowledgeMetrics),
		mockConversationRepository: mockConversationRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func TestWhenLabellingConversation(t *testing.T) {
	t.Parallel()

	t.Run("Given a named conversation is read back", func(t *testing.T) {
		t.Parallel()

		conversation := &models.Conversation{Id: "c1", Name: lo.ToPtr("Audit")}

		initSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
		}

		initMissingSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "ghost").Return(nil, nil)
		}

		t.Run("Should return its current label", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			name, err := suite.sut.GetName(context.Background(), "c1")

			assert.NoError(t, err)
			assert.Equal(t, "Audit", *name)
		})

		t.Run("Should return nil for an unknown conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initMissingSuite(suite)

			name, err := suite.sut.GetName(context.Background(), "ghost")

			assert.NoError(t, err)
			assert.Nil(t, name)
		})
	})

	t.Run("Given the conversation does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "ghost").Return(nil, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), "ghost", "Audit")

			assert.Error(t, err)
		})
	})

	t.Run("Given the name is free", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndName(gomock.Any(), conversation.AgentId, "ERC20 mint audit", lo.ToPtr("c1")).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Conversation) (*models.Conversation, error) {
					assert.Equal(t, "ERC20 mint audit", *updated.Name)
					return updated, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should store it as is", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.SetAvailableName(context.Background(), "c1", "ERC20 mint audit")

			assert.NoError(t, err)
			assert.Equal(t, "ERC20 mint audit", *updated.Name)
		})
	})

	t.Run("Given the session already holds that very name", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").WithName("Audit").Build()

		initSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndName(gomock.Any(), conversation.AgentId, "Audit", lo.ToPtr("c1")).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Conversation) (*models.Conversation, error) {
					return updated, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should accept it, since the lookup excludes the session itself", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetName(context.Background(), "c1", "Audit")

			assert.NoError(t, err)
		})
	})

	t.Run("Given another session already holds the name", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenLabellingConversationTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndName(gomock.Any(), conversation.AgentId, "Audit", lo.ToPtr("c1")).
				Return(true, nil)
		}

		t.Run("Should suffix when the caller tolerates a collision", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndName(gomock.Any(), conversation.AgentId, "Audit 2", lo.ToPtr("c1")).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Conversation) (*models.Conversation, error) {
					assert.Equal(t, "Audit 2", *updated.Name)
					return updated, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())

			_, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

			assert.NoError(t, err)
		})

		t.Run("Should report it taken when the caller does not", func(t *testing.T) {
			t.Parallel()

			suite := WhenLabellingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetName(context.Background(), "c1", "Audit")

			assert.ErrorIs(t, err, interfaces.ErrConversationNameTaken)
		})
	})
}
