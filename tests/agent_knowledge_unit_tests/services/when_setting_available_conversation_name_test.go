package services_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var settingAvailableNameCallerId = strings.ToLower(common_tests.RandomEthAddress().Hex())

type WhenSettingAvailableConversationNameTestingSuite struct {
	sut *services.ConversationLabeler

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenSettingAvailableConversationNameBeforeEach(t *testing.T) *WhenSettingAvailableConversationNameTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	return &WhenSettingAvailableConversationNameTestingSuite{
		sut:                        services.NewConversationLabeler(mockConversationRepository, mockKnowledgeMetrics),
		mockConversationRepository: mockConversationRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func TestWhenSettingAvailableConversationName(t *testing.T) {
	t.Parallel()

	t.Run("Given a padded conversation id", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Conversation) (*models.Conversation, error) {
					return updated, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should look the conversation up by the trimmed id", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), " c1 ", "Audit")

			assert.NoError(t, err)
		})
	})

	t.Run("Given the conversation lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given the conversation does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "ghost").Return(nil, nil)
		}

		t.Run("Should return a not-found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), "ghost", "Audit")

			assert.ErrorContains(t, err, "not found")
		})
	})

	t.Run("Given the name lookup fails", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), conversation.AgentId, gomock.Nil(), "Audit", lo.ToPtr("c1")).
				Return(false, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given another conversation holds that name", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), conversation.AgentId, gomock.Nil(), "Audit", lo.ToPtr("c1")).
				Return(true, nil)
		}

		t.Run("Should suffix the name", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), conversation.AgentId, gomock.Nil(), "Audit 2", lo.ToPtr("c1")).
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

		t.Run("Given every suffix up to the attempt cap is taken too", func(t *testing.T) {
			t.Parallel()

			t.Run("Should report the name as taken", func(t *testing.T) {
				t.Parallel()

				suite := WhenSettingAvailableConversationNameBeforeEach(t)
				initSuite(suite)
				for attempt := 2; attempt <= 10; attempt++ {
					suite.mockConversationRepository.EXPECT().
						ExistsByAgentIdAndCallerIdAndName(
							gomock.Any(),
							conversation.AgentId,
							gomock.Nil(),
							fmt.Sprintf("Audit %d", attempt),
							lo.ToPtr("c1"),
						).
						Return(true, nil)
				}
				suite.mockConversationRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)

				_, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

				assert.ErrorIs(t, err, interfaces.ErrConversationNameTaken)
			})
		})
	})

	t.Run("Given saving the name fails", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), conversation.AgentId, gomock.Nil(), "Audit", lo.ToPtr("c1")).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db blew"))
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any()).Times(0)
		}

		t.Run("Should propagate the error without counting a naming", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given the name is free for that caller", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().
			WithId("c1").
			WithCallerId(settingAvailableNameCallerId).
			Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(
					gomock.Any(),
					conversation.AgentId,
					lo.ToPtr(settingAvailableNameCallerId),
					"ERC20 mint audit",
					lo.ToPtr("c1"),
				).
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

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.SetAvailableName(context.Background(), "c1", "ERC20 mint audit")

			assert.NoError(t, err)
			assert.Equal(t, "ERC20 mint audit", *updated.Name)
		})
	})

	t.Run("Given the session already holds that very name", func(t *testing.T) {
		t.Parallel()

		conversation := model_builders.NewConversationBuilder().WithId("c1").WithName("Audit").Build()

		initSuite := func(suite *WhenSettingAvailableConversationNameTestingSuite) {
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), conversation.AgentId, gomock.Nil(), "Audit", lo.ToPtr("c1")).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Conversation) (*models.Conversation, error) {
					return updated, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAvailableConversationNameBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.SetAvailableName(context.Background(), "c1", "Audit")

			assert.NoError(t, err)
			assert.Equal(t, "Audit", *updated.Name)
		})
	})
}
