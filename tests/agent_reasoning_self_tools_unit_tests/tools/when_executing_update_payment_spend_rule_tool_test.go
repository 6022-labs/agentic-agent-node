package tools_test

import (
	"context"
	"errors"
	"testing"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingUpdatePaymentSpendRuleToolTestingSuite struct {
	sut *tools.UpdatePaymentSpendRuleTool

	mockUpdatePaymentSpendRuleExecutor *interfaces_mocks.MockUpdatePaymentSpendRuleExecutorInterface
}

func WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(
	t *testing.T,
) *WhenExecutingUpdatePaymentSpendRuleToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdatePaymentSpendRuleExecutor := interfaces_mocks.NewMockUpdatePaymentSpendRuleExecutorInterface(
		mockController,
	)

	return &WhenExecutingUpdatePaymentSpendRuleToolTestingSuite{
		sut:                                tools.NewUpdatePaymentSpendRuleTool(mockUpdatePaymentSpendRuleExecutor),
		mockUpdatePaymentSpendRuleExecutor: mockUpdatePaymentSpendRuleExecutor,
	}
}

func TestWhenExecutingUpdatePaymentSpendRuleTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	id := uuid.New()
	ruleArgs := func() map[string]any {
		return map[string]any{
			"id":                  id.String(),
			"name":                "Only base, small",
			"maxAmountPerPayment": "1000",
			"config":              map[string]any{"network": "eip155:8453"},
		}
	}

	t.Run("Given the id argument is not a uuid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			args := ruleArgs()
			args["id"] = "nope"
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "uuid")
		})
	})

	t.Run("Given the windows argument is not an array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			args := ruleArgs()
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

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)

			args := ruleArgs()
			delete(args, "config")
			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.EqualError(t, err, `argument "config" is required`)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("not found"))

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})

	t.Run("Given valid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pair the route id with the rule body", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.UpdatePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Equal(t, id, request.Id)
					assert.Equal(t, "Only base, small", request.Name)
					assert.Equal(t, "1000", *request.MaxAmountPerPayment)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				})

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})

	})

	t.Run("Given the optional windows argument", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingUpdatePaymentSpendRuleToolTestingSuite) {
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.UpdatePaymentSpendRuleRequest,
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

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			initSuite(suite)

			args := ruleArgs()
			args["windows"] = []any{
				map[string]any{"windowSeconds": float64(86400), "maxAmount": "50000"},
				map[string]any{"windowSeconds": float64(604800), "maxAmount": "200000"},
			}

			out, err := suite.sut.Execute(ctx, args)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the windows argument is omitted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingUpdatePaymentSpendRuleToolTestingSuite) {
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.UpdatePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Nil(t, request.Windows)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				}).
				Times(1)
		}

		t.Run("Should clear the windows on the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, ruleArgs())

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given no per-payment cap", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingUpdatePaymentSpendRuleToolTestingSuite) {
			suite.mockUpdatePaymentSpendRuleExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *payment_requests.UpdatePaymentSpendRuleRequest,
				) (*payment_responses.PaymentSpendRuleResponse, error) {
					assert.Nil(t, request.MaxAmountPerPayment)
					return &payment_responses.PaymentSpendRuleResponse{}, nil
				}).
				Times(1)
		}

		t.Run("Should leave the cap unset on the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdatePaymentSpendRuleToolBeforeEach(t)
			initSuite(suite)

			args := ruleArgs()
			delete(args, "maxAmountPerPayment")

			out, err := suite.sut.Execute(ctx, args)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
