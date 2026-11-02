package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingSignTypedDataToolTestingSuite struct {
	sut *tools.SignTypedDataTool

	mockSignTypedDataExecutor *interfaces_mocks.MockSignTypedDataExecutorInterface
}

func WhenExecutingSignTypedDataToolBeforeEach(t *testing.T) *WhenExecutingSignTypedDataToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSignTypedDataExecutor := interfaces_mocks.NewMockSignTypedDataExecutorInterface(mockController)

	return &WhenExecutingSignTypedDataToolTestingSuite{
		sut:                       tools.NewSignTypedDataTool(mockSignTypedDataExecutor),
		mockSignTypedDataExecutor: mockSignTypedDataExecutor,
	}
}

func TestWhenExecutingSignTypedDataTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the typedData argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignTypedDataToolBeforeEach(t)
			suite.mockSignTypedDataExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)

			out, err := suite.sut.Execute(ctx, map[string]any{})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "typedData")
		})
	})

	t.Run("Given a typedData argument that is not typed data", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignTypedDataToolBeforeEach(t)
			suite.mockSignTypedDataExecutor.EXPECT().Execute(gomock.Any(), gomock.Any()).Times(0)

			out, err := suite.sut.Execute(ctx, map[string]any{
				"typedData": map[string]any{"primaryType": 42.0},
			})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "typedData")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingSignTypedDataToolTestingSuite) {
			suite.mockSignTypedDataExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("node wallet not bootstrapped"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignTypedDataToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, map[string]any{"typedData": map[string]any{}})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not bootstrapped")
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingSignTypedDataToolTestingSuite) {
			suite.mockSignTypedDataExecutor.EXPECT().
				Execute(ctx, gomock.Cond(func(request *payment_requests.SignTypedDataRequest) bool {
					return request.TypedData != nil && request.TypedData.PrimaryType == "Mail"
				})).
				Return(&payment_responses.TypedDataSignatureResponse{Signature: "0xdead"}, nil)
		}

		t.Run("Should return the signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSignTypedDataToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, map[string]any{
				"typedData": map[string]any{"primaryType": "Mail"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "0xdead", out["signature"])
		})
	})
}
