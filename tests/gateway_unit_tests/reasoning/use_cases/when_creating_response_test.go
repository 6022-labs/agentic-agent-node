package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
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

func userInputRequest(conversationId, content string) *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		PreviousResponseId: conversationId,
		Input: requests.ResponseInput{
			Messages: []requests.InputMessage{{Role: "user", Content: content}},
		},
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
			PreviousResponseId: "conv-1",
			Instructions:       "you are helpful",
			Input: requests.ResponseInput{
				Messages: []requests.InputMessage{{Role: "user", Content: "hi"}},
			},
		}

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, payload *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error) {
					assert.Equal(t, "conv-1", payload.PreviousResponseId)
					assert.Equal(t, "you are helpful", payload.Instructions)
					if assert.Len(t, payload.Input, 1) {
						assert.Equal(t, "message", payload.Input[0].Type)
						assert.Equal(t, "user", payload.Input[0].Role)
						assert.Equal(t, "hi", payload.Input[0].Content)
					}
					return &models.ExternalCreateResponseResultPayload{
						Id:     "conv-1",
						Object: "response",
						Status: "completed",
						Output: []models.ExternalOutputItemPayload{
							{
								Type: "message",
								Role: "assistant",
								Content: []models.ExternalOutputContentPayload{
									{Type: "output_text", Text: "hello back"},
								},
							},
						},
					}, nil
				})
		}

		t.Run(
			"Should forward the mapped payload and return the runtime's OpenAI response verbatim",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenCreatingResponseBeforeEach(t)
				initSuite(suite)

				resp, err := suite.sut.Execute(ctx, request)

				assert.NoError(t, err)
				assert.Equal(t, "response", resp.Object)
				assert.Equal(t, "completed", resp.Status)
				assert.Equal(t, "conv-1", resp.Id)
				if assert.Len(t, resp.Output, 1) && assert.Len(t, resp.Output[0].Content, 1) {
					assert.Equal(t, "output_text", resp.Output[0].Content[0].Type)
					assert.Equal(t, "hello back", resp.Output[0].Content[0].Text)
				}
			},
		)
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

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})
}
