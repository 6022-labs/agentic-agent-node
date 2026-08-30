package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	knowledge_services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingConversationTestingSuite struct {
	sut *use_cases.UpdateConversation

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockConversationLabeler    *knowledge_services_mocks.MockConversationLabelerInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenUpdatingConversationBeforeEach(t *testing.T) *WhenUpdatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockConversationLabeler := knowledge_services_mocks.NewMockConversationLabelerInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenUpdatingConversationTestingSuite{
		sut: use_cases.NewUpdateConversation(
			mockConversationRepository,
			mockConversationLabeler,
			mockAgentRepository,
		),
		mockConversationRepository: mockConversationRepository,
		mockConversationLabeler:    mockConversationLabeler,
		mockAgentRepository:        mockAgentRepository,
	}
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	t.Run("Given the conversation is not the agent's", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		foreign := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(uuid.New()).Build()

		initForeignSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(foreign, nil)
		}

		initMissingSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "ghost").Return(nil, nil)
		}

		t.Run("Should report it as not found rather than forbidden", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initForeignSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "c1",
				Name:           "mine",
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})

		t.Run("Should report an unknown id as not found", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initMissingSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "ghost",
				Name:           "x y",
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given another session already holds the name", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(agent.Id).Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "taken").
				Return(nil, interfaces.ErrConversationNameTaken)
		}

		t.Run("Should return a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "c1",
				Name:           "taken",
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given the agent owns the conversation", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(agent.Id).Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().GetById(gomock.Any(), "c1").Return(conversation, nil)
		}

		t.Run("Should rename it", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "Solidity audits").
				Return(model_builders.NewConversationBuilder().
					WithId("c1").
					WithAgentId(agent.Id).
					WithName("Solidity audits").
					Build(), nil)

			resp, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "c1",
				Name:           "Solidity audits",
			})

			assert.NoError(t, err)
			assert.Equal(t, "Solidity audits", *resp.Name)
		})

		t.Run("Should accept renaming a session to the name it already holds", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "same").
				Return(conversation, nil)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationId: "c1",
				Name:           "same",
			})

			assert.NoError(t, err)
		})
	})
}
