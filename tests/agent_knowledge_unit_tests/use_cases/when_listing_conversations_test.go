package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	listingCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	listingCallerCtx = helpers.ContextWithCallerId(context.Background(), listingCallerId)
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

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.ListConversationsRequest{})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
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

			resp, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Empty(t, resp.Conversations)
			assert.Equal(t, uint(1), resp.Page)
		})
	})

	t.Run("Given the caller does not own the agent", func(t *testing.T) {
		t.Parallel()

		callerId := listingCallerId

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), agent.Id, &callerId).
				Return(uint(1), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), agent.Id, &callerId, lo.ToPtr(0), lo.ToPtr(10)).
				Return([]models.Conversation{
					*model_builders.NewConversationBuilder().WithId("mine").WithCallerId(callerId).Build(),
				}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), []string{"mine"}).
				Return(map[string]int64{}, nil)
		}

		t.Run("Should list only that caller's sessions", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 1)
			assert.Equal(t, "mine", resp.Conversations[0].ConversationId)
		})
	})

	t.Run("Given the caller owns the agent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(listingCallerId)).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), agent.Id, gomock.Nil()).
				Return(uint(2), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), agent.Id, gomock.Nil(), lo.ToPtr(0), lo.ToPtr(10)).
				Return([]models.Conversation{
					*model_builders.NewConversationBuilder().WithId("mine").WithCallerId(listingCallerId).Build(),
					*model_builders.NewConversationBuilder().WithId("owner").Build(),
				}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), []string{"mine", "owner"}).
				Return(map[string]int64{}, nil)
		}

		t.Run("Should list every session of the agent unscoped", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 2)
		})
	})

	t.Run("Given counting the sessions fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId)).
				Return(uint(0), errors.New("db blew"))
		}

		t.Run("Should return an InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})

	t.Run("Given listing the sessions fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId)).
				Return(uint(1), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})

	t.Run("Given counting the messages fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId)).
				Return(uint(1), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId), gomock.Any(), gomock.Any()).
				Return([]models.Conversation{*model_builders.NewConversationBuilder().WithId("a").Build()}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), []string{"a"}).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})

	t.Run("Given a page of sessions", func(t *testing.T) {
		t.Parallel()

		createdAt := time.Now().UTC().Truncate(time.Microsecond)

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockConversationRepository.EXPECT().
				CountByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId)).
				Return(uint(2), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId), lo.ToPtr(0), lo.ToPtr(10)).
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

			resp, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

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
				CountByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId)).
				Return(uint(1), nil)
			suite.mockConversationRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any(), lo.ToPtr(listingCallerId), gomock.Any(), gomock.Any()).
				Return([]models.Conversation{
					*model_builders.NewConversationBuilder().WithId("empty").WithName("Solidity audits").Build(),
				}, nil)
			suite.mockConversationMessageRepository.EXPECT().
				CountByConversationIds(gomock.Any(), gomock.Any()).
				Return(map[string]int64{}, nil)
		}

		t.Run("Should list it with a zero message count", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(listingCallerCtx, &requests.ListConversationsRequest{})

			assert.NoError(t, err)
			assert.Len(t, resp.Conversations, 1)
			assert.Equal(t, int64(0), resp.Conversations[0].MessageCount)
		})
	})
}
