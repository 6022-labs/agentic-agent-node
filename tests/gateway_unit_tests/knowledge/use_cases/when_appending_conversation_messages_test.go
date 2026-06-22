package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockClient *interfaces_mocks.MockConversationHistoryClientInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationHistoryClientInterface(mockController)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:        use_cases.NewAppendConversationMessages(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	request := &requests.AppendConversationMessagesRequest{
		Messages: []requests.AppendConversationMessageInput{
			{Role: enums.MessageRoleUser, Content: "fake memory", Timestamp: now},
		},
	}

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), "conv-1", &requests.AppendConversationMessagesRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the client accepts the messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, payload *models.ExternalAppendConversationMessagesPayload) (*models.ExternalAppendConversationMessagesResultPayload, error) {
					assert.Len(t, payload.Messages, 1)
					assert.Equal(t, "user", payload.Messages[0].Role)
					assert.Equal(t, "fake memory", payload.Messages[0].Content)
					return &models.ExternalAppendConversationMessagesResultPayload{
						Inserted: []models.ExternalConversationMessage{
							{Id: "m1", ConversationId: conversationId, Role: "user", Content: lo.ToPtr("fake memory"), Timestamp: now},
						},
					}, nil
				})
		}

		t.Run("Should translate the request and return the inserted messages", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), conversationId, request)

			assert.NoError(t, err)
			assert.Len(t, resp.Inserted, 1)
			assert.Equal(t, "m1", resp.Inserted[0].Id)
			assert.Equal(t, "fake memory", lo.FromPtr(resp.Inserted[0].Content))
		})
	})

	t.Run("Given the client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().AppendMessages(gomock.Any(), conversationId, gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return Unavailable carrying the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId, request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})
}
