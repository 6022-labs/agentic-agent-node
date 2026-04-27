package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)

	sut := use_cases.NewListConversationMessages(mockRepository)

	return &WhenListingConversationMessagesTestingSuite{
		sut: sut,

		mockRepository: mockRepository,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	t.Run("Given an empty conversationId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)

			res, err := suite.sut.Execute("   ")

			assert.NoError(t, err)
			assert.Nil(t, res)
		})
	})

	t.Run("Given the repository returns rows", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		expected := []models.ConversationMessage{
			{Id: uuid.New(), ConversationId: "conv", Timestamp: now},
			{Id: uuid.New(), ConversationId: "conv", Timestamp: now.Add(time.Second)},
		}
		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().ListByConversation("conv").Return(expected, nil)
		}

		t.Run("Should return them as-is", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute("conv")

			assert.NoError(t, err)
			assert.Equal(t, expected, res)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockRepository.EXPECT().ListByConversation("conv").Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute("conv")

			assert.ErrorContains(t, err, "db blew")
		})
	})
}
