package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCreatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.CreatePaymentSpendRuleTool

	mockCreatePaymentSpendRuleExecutor *interfaces_mocks.MockCreatePaymentSpendRuleExecutorInterface
}

func WhenExecutingCreatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingCreatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreatePaymentSpendRuleExecutor := interfaces_mocks.NewMockCreatePaymentSpendRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingCreatePaymentSpendRuleToolTestingSuite{
		sut:                                tools.NewCreatePaymentSpendRuleTool(mockCreatePaymentSpendRuleExecutor),
		mockCreatePaymentSpendRuleExecutor: mockCreatePaymentSpendRuleExecutor,
	}
}

func TestWhenExecutingCreatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	validArgs := func() map[string]any {
		return map[string]any{
			"name":                "Only base, small",
			"maxAmountPerPayment": "1000",
			"config":              map[string]any{"network": "eip155:8453"},
		}
	}

	t.Run("Given the windows argument is not an array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			args := validArgs()
			args["windows"] = "daily"
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "must be a json array")
		})
	})

	t.Run("Given a required argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			args := validArgs()
			delete(args, "config")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.EqualError(t, err, `argument "config" is required`)
		})
	})

	t.Run("Given a per-payment cap that is not a string", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)

			args := validArgs()
			args["maxAmountPerPayment"] = map[string]any{"amount": "1000"}
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.EqualError(t, err, `argument "maxAmountPerPayment" must be a string`)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map them onto the save request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.SavePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Equal(t, "Only base, small", request.Name)
					assert.Equal(t, "1000", *request.MaxAmountPerPayment)
					assert.Contains(t, string(request.Config), "eip155:8453")
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, validArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})

	})

	t.Run("Given the optional windows argument", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingCreatePaymentSpendRuleToolTestingSuite) {
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.SavePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Equal(t, []payment_requests.SavePaymentSpendRuleWindowRequest{
						{WindowSeconds: 86400, MaxAmount: "50000"},
						{WindowSeconds: 604800, MaxAmount: "200000"},
					}, request.Windows)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				}).
				Times(1)
		}

		t.Run("Should carry every window onto the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			initSuite(suite)

			args := validArgs()
			args["windows"] = []any{
				map[string]any{"windowSeconds": float64(86400), "maxAmount": "50000"},
				map[string]any{"windowSeconds": float64(604800), "maxAmount": "200000"},
			}

			out, err := suite.sut.Execute(ctx, args)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given no per-payment cap", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingCreatePaymentSpendRuleToolTestingSuite) {
			suite.mockCreatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.SavePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Nil(t, request.MaxAmountPerPayment)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				}).
				Times(1)
		}

		t.Run("Should leave the cap unset on the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreatePaymentSpendRuleToolBeforeEach(t)
			initSuite(suite)

			args := validArgs()
			delete(args, "maxAmountPerPayment")

			out, err := suite.sut.Execute(ctx, args)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
