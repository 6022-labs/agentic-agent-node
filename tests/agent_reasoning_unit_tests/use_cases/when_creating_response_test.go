package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseTestingSuite struct {
	sut *use_cases.CreateResponse

	mockResponseDriver *interfaces_mocks.MockResponseDriverInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockResponseDriver := interfaces_mocks.NewMockResponseDriverInterface(mockController)

	sut := use_cases.NewCreateResponse(mockResponseDriver)

	return &WhenCreatingResponseTestingSuite{
		sut:                sut,
		mockResponseDriver: mockResponseDriver,
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given the response driver returns a persisted assistant message", func(t *testing.T) {
		t.Parallel()

		expectedFinal := "hi there"

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(ctx, conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, inputs []knowledge_interfaces.ConversationMessage) (*knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, inputs, 1) {
						assert.Equal(t, "user", inputs[0].Role)
						assert.Equal(t, "hello", inputs[0].Content)
					}
					return &knowledge_interfaces.ConversationMessage{
						Role:      "assistant",
						Content:   expectedFinal,
						Timestamp: now,
					}, nil
				})
		}

		t.Run("Should return the assistant message in the response DTO", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: conversationId,
				Inputs: []requests.CreateResponseRequestInput{
					{Role: "user", Content: "hello", Timestamp: now},
				},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Message) {
				assert.Equal(t, "assistant", result.Message.Role)
				assert.Equal(t, expectedFinal, result.Message.Content)
			}
		})
	})

	t.Run("Given the response driver returns nil", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(ctx, conversationId, gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return an empty response without an assistant message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: conversationId,
				Inputs: []requests.CreateResponseRequestInput{
					{Role: "user", Content: "hello", Timestamp: now},
				},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.Message)
			}
		})
	})

	t.Run("Given the response driver fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(ctx, conversationId, gomock.Any()).
				Return(nil, errors.New("driver broke"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: conversationId,
				Inputs: []requests.CreateResponseRequestInput{
					{Role: "user", Content: "hello", Timestamp: now},
				},
			})

			assert.Nil(t, result)
			assert.ErrorContains(t, err, "driver broke")
		})
	})
}
