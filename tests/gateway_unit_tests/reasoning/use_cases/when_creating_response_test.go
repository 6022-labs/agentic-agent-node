package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseTestingSuite struct {
	sut *use_cases.CreateResponse

	mockClient *interfaces_mocks.MockResponseClientInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockResponseClientInterface(mockController)

	sut := use_cases.NewCreateResponse(mockClient)

	return &WhenCreatingResponseTestingSuite{
		sut:        sut,
		mockClient: mockClient,
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given a request that fails validation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)

			_, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a valid request the runtime accepts", func(t *testing.T) {
		t.Parallel()

		request := &requests.CreateResponseRequest{
			ConversationId: "conv-1",
			Platform:       enums.ConversationPlatformSlack,
			Prompt:         "you are helpful",
			Input: []*requests.InputMessageRequest{
				{Id: "m1", Author: "alice", Role: enums.MessageRoleUser, Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, payload *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error) {
					assert.Equal(t, "conv-1", payload.ConversationId)
					if assert.Len(t, payload.Inputs, 1) {
						assert.Equal(t, "user", payload.Inputs[0].Role)
						assert.Equal(t, "hi", payload.Inputs[0].Content)
						assert.Equal(t, now, payload.Inputs[0].Timestamp)
					}
					return &models.ExternalCreateResponseResultPayload{
						Message: &models.ExternalCreateResponseMessagePayload{
							Role:      "assistant",
							Content:   "hello back",
							Timestamp: now,
						},
					}, nil
				})
		}

		t.Run("Should return Ok forwarding inputs and surfacing the assistant message in Response.Output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			assert.True(t, resp.IsFinal)
			if assert.Len(t, resp.Output, 1) {
				assert.Equal(t, "hello back", resp.Output[0].Content)
				assert.Equal(t, now, resp.Output[0].Timestamp)
			}
		})
	})

	t.Run("Given the runtime returns no assistant message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(&models.ExternalCreateResponseResultPayload{}, nil)
		}

		t.Run("Should return Ok with IsFinal=false and no Output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: "conv-1",
				Prompt:         "You are a helpful assistant.",
				Input: []*requests.InputMessageRequest{
					{Id: "m1", Role: enums.MessageRoleUser, Content: "hi", Timestamp: now},
				},
			})

			assert.NoError(t, err)
			assert.False(t, resp.IsFinal)
			assert.Empty(t, resp.Output)
		})
	})

	t.Run("Given the runtime client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return Unavailable carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: "conv-1",
				Prompt:         "You are a helpful assistant.",
				Input: []*requests.InputMessageRequest{
					{Id: "m1", Role: enums.MessageRoleUser, Content: "hi", Timestamp: now},
				},
			})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})
}
