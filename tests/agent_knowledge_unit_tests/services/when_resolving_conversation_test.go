package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	resolverAlice = strings.ToLower(common_tests.RandomEthAddress().Hex())
	resolverBob   = strings.ToLower(common_tests.RandomEthAddress().Hex())
)

type WhenResolvingConversationTestingSuite struct {
	sut *services.ConversationResolver

	agent *identity_models.Agent

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenResolvingConversationBeforeEach(t *testing.T) *WhenResolvingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	return &WhenResolvingConversationTestingSuite{
		sut: services.NewConversationResolver(
			mockConversationRepository,
			mockAgentRepository,
			mockKnowledgeMetrics,
		),
		agent:                      model_builders.NewAgentBuilder().Build(),
		mockConversationRepository: mockConversationRepository,
		mockAgentRepository:        mockAgentRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func (s *WhenResolvingConversationTestingSuite) expectOwnedAgent(owner string) {
	s.agent = model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(owner)).Build()
	s.expectAgent()
}

func (s *WhenResolvingConversationTestingSuite) expectAgent() {
	s.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(s.agent, nil)
}

func (s *WhenResolvingConversationTestingSuite) expectEnsure(conversationId string, created bool) {
	callerId := resolverAlice
	s.mockConversationRepository.EXPECT().
		Ensure(gomock.Any(), &models.Conversation{
			Id:       conversationId,
			AgentId:  s.agent.Id,
			CallerId: &callerId,
		}).
		Return(created, nil)
}

func (s *WhenResolvingConversationTestingSuite) expectStored(conversationId string, callerId *string) {
	builder := model_builders.NewConversationBuilder().WithId(conversationId).WithAgentId(s.agent.Id)
	if callerId != nil {
		builder = builder.WithCallerId(*callerId)
	}
	s.mockConversationRepository.EXPECT().GetById(gomock.Any(), conversationId).Return(builder.Build(), nil)
}

func TestWhenResolvingConversation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	alice := resolverAlice
	bob := resolverBob

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db blew"))

			_, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrAgentNotMinted", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)

			_, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.ErrorIs(t, err, interfaces.ErrAgentNotMinted)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrAgentNotMinted", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build(), nil)

			_, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.ErrorIs(t, err, interfaces.ErrAgentNotMinted)
		})
	})

	t.Run("Given no conversation key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should claim a new session", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			var ensuredId string
			suite.expectAgent()
			suite.mockConversationRepository.EXPECT().
				Ensure(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conversation *models.Conversation) (bool, error) {
					ensuredId = conversation.Id
					assert.Equal(t, suite.agent.Id, conversation.AgentId)
					assert.Equal(t, alice, *conversation.CallerId)
					return true, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())

			conversationId, err := suite.sut.Resolve(ctx, "  ", alice)

			assert.NoError(t, err)
			assert.Equal(t, ensuredId, conversationId)
			_, parseErr := uuid.Parse(conversationId)
			assert.NoError(t, parseErr)
		})
	})

	t.Run("Given a padded conversation key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should claim it trimmed", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.expectAgent()
			suite.expectEnsure("c1", true)
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())

			conversationId, err := suite.sut.Resolve(ctx, " c1 ", alice)

			assert.NoError(t, err)
			assert.Equal(t, "c1", conversationId)
		})
	})

	t.Run("Given a key nobody holds yet", func(t *testing.T) {
		t.Parallel()

		t.Run("Given claiming it fails", func(t *testing.T) {
			t.Parallel()

			t.Run("Should propagate the error", func(t *testing.T) {
				t.Parallel()

				suite := WhenResolvingConversationBeforeEach(t)
				suite.expectAgent()
				suite.mockConversationRepository.EXPECT().
					Ensure(gomock.Any(), gomock.Any()).
					Return(false, errors.New("db blew"))

				_, err := suite.sut.Resolve(ctx, "c1", alice)

				assert.ErrorContains(t, err, "db blew")
			})
		})

		t.Run("Should claim it", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.expectAgent()
			suite.expectEnsure("c1", true)
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())

			conversationId, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.NoError(t, err)
			assert.Equal(t, "c1", conversationId)
		})
	})

	t.Run("Given a key the caller already holds", func(t *testing.T) {
		t.Parallel()

		t.Run("Given the conversation lookup fails", func(t *testing.T) {
			t.Parallel()

			t.Run("Should propagate the error", func(t *testing.T) {
				t.Parallel()

				suite := WhenResolvingConversationBeforeEach(t)
				suite.expectAgent()
				suite.expectEnsure("c1", false)
				suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(nil, errors.New("db blew"))

				_, err := suite.sut.Resolve(ctx, "c1", alice)

				assert.ErrorContains(t, err, "db blew")
			})
		})

		t.Run("Should resume it", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.expectAgent()
			suite.expectEnsure("c1", false)
			suite.expectStored("c1", &alice)

			conversationId, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.NoError(t, err)
			assert.Equal(t, "c1", conversationId)
		})
	})

	t.Run("Given a key another caller holds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrConversationUnavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.expectAgent()
			suite.expectEnsure("c1", false)
			suite.expectStored("c1", &bob)

			_, err := suite.sut.Resolve(ctx, "c1", alice)

			assert.ErrorIs(t, err, interfaces.ErrConversationUnavailable)
		})
	})

	t.Run("Given a key held by a caller-less session", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrConversationUnavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingConversationBeforeEach(t)
			suite.expectAgent()
			suite.expectEnsure("legacy", false)
			suite.expectStored("legacy", nil)

			_, err := suite.sut.Resolve(ctx, "legacy", alice)

			assert.ErrorIs(t, err, interfaces.ErrConversationUnavailable)
		})
	})

	t.Run("Given the caller owns the agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Given a key another caller holds", func(t *testing.T) {
			t.Parallel()

			t.Run("Should resume it", func(t *testing.T) {
				t.Parallel()

				suite := WhenResolvingConversationBeforeEach(t)
				suite.expectOwnedAgent(alice)
				suite.expectEnsure("c1", false)
				suite.expectStored("c1", &bob)

				conversationId, err := suite.sut.Resolve(ctx, "c1", alice)

				assert.NoError(t, err)
				assert.Equal(t, "c1", conversationId)
			})
		})

		t.Run("Given a key held by a caller-less session", func(t *testing.T) {
			t.Parallel()

			t.Run("Should resume it", func(t *testing.T) {
				t.Parallel()

				suite := WhenResolvingConversationBeforeEach(t)
				suite.expectOwnedAgent(alice)
				suite.expectEnsure("legacy", false)
				suite.expectStored("legacy", nil)

				conversationId, err := suite.sut.Resolve(ctx, "legacy", alice)

				assert.NoError(t, err)
				assert.Equal(t, "legacy", conversationId)
			})
		})

		t.Run("Given a key that belongs to another agent", func(t *testing.T) {
			t.Parallel()

			t.Run("Should return ErrConversationUnavailable", func(t *testing.T) {
				t.Parallel()

				suite := WhenResolvingConversationBeforeEach(t)
				suite.expectOwnedAgent(alice)
				suite.expectEnsure("c1", false)
				suite.mockConversationRepository.EXPECT().
					GetById(gomock.Any(), "c1").
					Return(model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(uuid.New()).
						WithCallerId(bob).
						Build(), nil)

				_, err := suite.sut.Resolve(ctx, "c1", alice)

				assert.ErrorIs(t, err, interfaces.ErrConversationUnavailable)
			})
		})
	})
}
