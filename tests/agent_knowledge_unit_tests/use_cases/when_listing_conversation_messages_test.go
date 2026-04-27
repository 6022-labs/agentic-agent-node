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

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	t.Run("Given an empty conversationId", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		uc := use_cases.NewListConversationMessages(repo)

		t.Run("Should return nil without touching the repository", func(t *testing.T) {
			res, err := uc.Execute("   ")
			assert.NoError(t, err)
			assert.Nil(t, res)
		})
	})

	t.Run("Given the repository returns rows", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()
		expected := []models.ConversationMessage{
			{Id: uuid.New(), ConversationId: "conv", Timestamp: now},
			{Id: uuid.New(), ConversationId: "conv", Timestamp: now.Add(time.Second)},
		}
		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		repo.EXPECT().ListByConversation("conv").Return(expected, nil)

		uc := use_cases.NewListConversationMessages(repo)

		t.Run("Should return them as-is", func(t *testing.T) {
			res, err := uc.Execute("conv")
			assert.NoError(t, err)
			assert.Equal(t, expected, res)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		repo.EXPECT().ListByConversation("conv").Return(nil, errors.New("db blew"))

		uc := use_cases.NewListConversationMessages(repo)

		t.Run("Should propagate the error", func(t *testing.T) {
			_, err := uc.Execute("conv")
			assert.ErrorContains(t, err, "db blew")
		})
	})
}
