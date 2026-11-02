package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentSpendPolicyTestingSuite struct {
	sut *services.HttpPaymentSpendPolicyClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenUpdatingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenUpdatingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenUpdatingPaymentSpendPolicyTestingSuite{
		sut:        services.NewHttpPaymentSpendPolicyClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentSpendPolicyWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(
				context.Background(),
				&models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: []string{"0xabc.*"}},
			)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentSpendPolicyWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PutInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"payToPatterns","message":"patterns cannot be empty"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Update(
				context.Background(),
				&models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: []string{" "}},
			)

			assert.Nil(t, payload)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "payToPatterns", validationErr.Field)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentSpendPolicyWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Update(
				context.Background(),
				&models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: []string{"0xabc.*"}},
			)

			assert.Nil(t, payload)
			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the payees", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentSpendPolicyWithBodyWithResponse(
					gomock.Any(),
					"application/json",
					gomock.Cond(func(body io.Reader) bool {
						var sent models.ExternalUpdatePaymentSpendPolicyPayload
						return json.NewDecoder(body).Decode(&sent) == nil &&
							len(sent.PayToPatterns) == 1 && sent.PayToPatterns[0] == "0xabc.*"
					}),
				).
				Return(&clients.PutInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"id":"a","agentId":"b","payToPatterns":["0xabc.*"]}`),
				}, nil).
				Times(1)
		}

		t.Run("Should send the payees as the json body", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(
				context.Background(),
				&models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: []string{"0xabc.*"}},
			)

			assert.NoError(t, err)
		})

		t.Run("Should decode the updated policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Update(
				context.Background(),
				&models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: []string{"0xabc.*"}},
			)

			assert.NoError(t, err)
			assert.Equal(t, "a", payload.Id)
			assert.Equal(t, []string{"0xabc.*"}, payload.PayToPatterns)
		})
	})
}
