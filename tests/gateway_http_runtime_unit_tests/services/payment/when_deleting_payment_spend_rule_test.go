package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentSpendRuleTestingSuite struct {
	sut *services.HttpPaymentSpendRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingPaymentSpendRuleBeforeEach(t *testing.T) *WhenDeletingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingPaymentSpendRuleTestingSuite{
		sut:        services.NewHttpPaymentSpendRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), "r1")

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), "r1")

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), "r1")

			var target *apperrors.NotFoundError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), "r1").
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), "r1")

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the rule is deleted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), "r1").
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
				}, nil)
		}

		t.Run("Should delete without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			assert.NoError(t, suite.sut.Delete(context.Background(), "r1"))
		})
	})
}
