package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationsTestingSuite struct {
	sut *use_cases.ListConversations

	mockConversationRepository        *knowledge_mocks.MockConversationRepositoryInterface
	mockConversationMessageRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
	mockAgentRepository               *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenListingConversationsBeforeEach(t *testing.T) *WhenListingConversationsTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockConversationMessageRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenListingConversationsTestingSuite{
		sut: use_cases.NewListConversations(
			mockConversationRepository,
			mockConversationMessageRepository,
			mockAgentRepository,
		),
		mockConversationRepository:        mockConversationRepository,
		mockConversationMessageRepository: mockConversationMessageRepository,
		mockAgentRepository:               mockAgentRepository,
	}
}

func TestWhenListingConversations(t *testing.T) {
	t.Parallel()

	t.Run("Given a page of sessions", func(t *testing.T) {
		t.Parallel()

		createdAt := time.Now().UTC().Truncate(time.Microsecond)

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any()).
				Return(uint(2), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(0), lo.ToPtr(10)).
				Return([]models.Conversation{
					*model_builders.NewConversationBuilder().
						WithId("a").
						WithName("ERC20 mint audit").
						WithCreatedAt(createdAt).
						Build(),
					*model_builders.NewConversationBuilder().WithId("b").WithCreatedAt(createdAt).Build(),
				}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), []string{"a", "b"}).
				Return(map[string]int64{"a": 2, "b": 1}, nil)
		}

		t.Run("Should decorate each session with its message count", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Equal(t, uint(1), resp.Page)
			assert.Equal(t, uint(10), resp.ElementsPerPage)
			assert.Equal(t, uint(2), resp.TotalElements)
			assert.Len(t, resp.Conversations, 2)
			assert.Equal(t, "a", resp.Conversations[0].ConversationId)
			assert.Equal(t, "ERC20 mint audit", *resp.Conversations[0].Name)
			assert.Equal(t, int64(2), resp.Conversations[0].MessageCount)
			assert.Equal(t, createdAt, resp.Conversations[0].CreatedAt)
			assert.Nil(t, resp.Conversations[1].Name)
			assert.Equal(t, int64(1), resp.Conversations[1].MessageCount)
		})
	})

	t.Run("Given a session that holds no message yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any()).
				Return(uint(1), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]models.Conversation{
					*model_builders.NewConversationBuilder().WithId("empty").WithName("Solidity audits").Build(),
				}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), gomock.Any()).
				Return(map[string]int64{}, nil)
		}

		t.Run("Should still list it, so a freshly created session is visible", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 1)
			assert.Equal(t, int64(0), resp.Conversations[0].MessageCount)
		})
	})

	t.Run("Given an out-of-range page size", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().CountByAgentId(gomock.Any(), gomock.Any()).Return(uint(0), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(0), lo.ToPtr(100)).
				Return(nil, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), gomock.Any()).
				Return(map[string]int64{}, nil)
		}

		t.Run("Should clamp it before hitting the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{
				Page:            lo.ToPtr(uint(0)),
				ElementsPerPage: lo.ToPtr(uint(5000)),
			})

			assert.NoError(t, err)
			assert.Equal(t, uint(100), resp.ElementsPerPage)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return an empty page", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Empty(t, resp.Conversations)
			assert.Equal(t, uint(1), resp.Page)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any()).
				Return(uint(0), errors.New("db blew"))
		}

		t.Run("Should return InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})
}
