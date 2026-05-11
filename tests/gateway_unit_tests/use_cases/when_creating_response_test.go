package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
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
				DoAndReturn(func(_ context.Context, payload *runtime_models.ExternalCreateResponsePayload) (*runtime_models.ExternalCreateResponseResultPayload, error) {
					assert.Equal(t, "conv-1", payload.ConversationId)
					if assert.Len(t, payload.Inputs, 1) {
						assert.Equal(t, "user", payload.Inputs[0].Role)
						assert.Equal(t, "hi", payload.Inputs[0].Content)
						assert.Equal(t, now, payload.Inputs[0].Timestamp)
					}
					return &runtime_models.ExternalCreateResponseResultPayload{
						Message: &runtime_models.ExternalCreateResponseMessagePayload{
							Role:      "assistant",
							Content:   "hello back",
							Timestamp: now,
						},
					}, nil
				})
		}

		t.Run("Should forward inputs and surface the assistant message in Response.Output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.True(t, result.IsFinal)
				if assert.Len(t, result.Output, 1) {
					assert.Equal(t, "hello back", result.Output[0].Content)
					assert.Equal(t, now, result.Output[0].Timestamp)
				}
			}
		})
	})

	t.Run("Given the runtime returns no assistant message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(&runtime_models.ExternalCreateResponseResultPayload{}, nil)
		}

		t.Run("Should return Response with IsFinal=false and no Output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: "conv-1",
				Input: []*requests.InputMessageRequest{
					{Id: "m1", Role: enums.MessageRoleUser, Content: "hi", Timestamp: now},
				},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.False(t, result.IsFinal)
				assert.Empty(t, result.Output)
			}
		})
	})

	t.Run("Given the runtime client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				ConversationId: "conv-1",
				Input: []*requests.InputMessageRequest{
					{Id: "m1", Role: enums.MessageRoleUser, Content: "hi", Timestamp: now},
				},
			})

			assert.Nil(t, result)
			assert.ErrorContains(t, err, "failed to create response")
			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
