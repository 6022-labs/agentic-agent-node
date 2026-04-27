package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given an empty conversationId", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		uc := use_cases.NewAppendConversationMessages(repo)

		t.Run("Should no-op without touching the repository", func(t *testing.T) {
			res, err := uc.Execute("   ", []models.ConversationMessage{{Author: "a", Content: "x", Timestamp: now}})
			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given an empty messages slice", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		uc := use_cases.NewAppendConversationMessages(repo)

		t.Run("Should no-op", func(t *testing.T) {
			res, err := uc.Execute(conversationId, nil)
			assert.NoError(t, err)
			assert.Empty(t, res)
		})
	})

	t.Run("Given a list of messages", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		repo.EXPECT().AppendMessages(gomock.Any()).DoAndReturn(func(messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
			assert.Len(t, messages, 2)
			for _, m := range messages {
				assert.Equal(t, conversationId, m.ConversationId)
			}
			return messages, nil
		})

		uc := use_cases.NewAppendConversationMessages(repo)

		t.Run("Should stamp conversationId on every row before delegating", func(t *testing.T) {
			input := []models.ConversationMessage{
				{Author: "alice", Role: "user", Content: "hi", Timestamp: now},
				{Author: "bob", Role: "assistant", Content: "hello", Timestamp: now},
			}
			res, err := uc.Execute(conversationId, input)
			assert.NoError(t, err)
			assert.Len(t, res, 2)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := knowledge_mocks.NewMockConversationMessageRepositoryInterface(ctrl)
		repo.EXPECT().AppendMessages(gomock.Any()).Return(nil, errors.New("db down"))

		uc := use_cases.NewAppendConversationMessages(repo)

		t.Run("Should propagate the error", func(t *testing.T) {
			_, err := uc.Execute(conversationId, []models.ConversationMessage{{Author: "a", Content: "x", Timestamp: now}})
			assert.ErrorContains(t, err, "db down")
		})
	})
}
