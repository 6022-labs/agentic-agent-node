package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
	use_cases_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/use_cases_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSendingA2AMessageTestingSuite struct {
	sut *use_cases.SendA2AMessage

	mockCreateResponse *use_cases_mocks.MockCreateResponseInterface
}

func WhenSendingA2AMessageBeforeEach(t *testing.T) *WhenSendingA2AMessageTestingSuite {
	mockController := gomock.NewController(t)

	mockCreateResponse := use_cases_mocks.NewMockCreateResponseInterface(mockController)

	sut := use_cases.NewSendA2AMessage(mockCreateResponse)

	return &WhenSendingA2AMessageTestingSuite{
		sut:                sut,
		mockCreateResponse: mockCreateResponse,
	}
}

func a2aTextRequest(method requests.A2AMethod, text, contextId string) *requests.A2ARequest {
	return &requests.A2ARequest{
		JsonRpc: "2.0",
		Id:      json.RawMessage(`1`),
		Method:  method,
		Params: requests.A2AMessageSendParams{
			Message: requests.A2AMessage{
				Role:      "ROLE_USER",
				Parts:     []requests.A2APart{{Text: text}},
				ContextId: contextId,
			},
		},
	}
}

func TestWhenSendingA2AMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling inference", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)

			_, err := suite.sut.Execute(ctx, a2aTextRequest("GetTask", "hi", ""))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a message with no text", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "   ", ""))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a valid SendMessage with a contextId", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *requests.CreateResponseRequest) (*responses.Response, error) {
					if assert.Len(t, request.Input.Messages, 1) {
						assert.Equal(t, "user", request.Input.Messages[0].Role)
						assert.Equal(t, "hello", request.Input.Messages[0].Content)
					}
					if assert.NotNil(t, request.Conversation) {
						assert.Equal(t, "ctx-1", request.Conversation.Id)
					}
					return &responses.Response{
						Output: []responses.OutputItem{
							{Role: "assistant", Content: []responses.OutputContent{{Type: "output_text", Text: "hi there"}}},
						},
						Conversation: &responses.ConversationRef{Id: "ctx-1"},
					}, nil
				})
		}

		t.Run("Should map to a Responses turn and return a ProtoJSON agent message", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", "ctx-1"))

			assert.NoError(t, err)
			assert.Equal(t, "2.0", resp.JsonRpc)
			if assert.NotNil(t, resp.Result) && assert.NotNil(t, resp.Result.Message) {
				message := resp.Result.Message
				assert.Equal(t, "ROLE_AGENT", message.Role)
				assert.Equal(t, "ctx-1", message.ContextId)
				assert.NotEmpty(t, message.MessageId)
				if assert.Len(t, message.Parts, 1) {
					assert.Equal(t, "hi there", message.Parts[0].Text)
				}
			}
		})
	})

	t.Run("Given inference fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should propagate the error so the edge maps it to >=400", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})
}
