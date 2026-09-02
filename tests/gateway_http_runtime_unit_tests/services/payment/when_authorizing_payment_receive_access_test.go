package services_test

import (
	"context"
	"encoding/json"
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

const gatePayerAddress = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

type WhenAuthorizingPaymentReceiveAccessTestingSuite struct {
	sut *services.HttpPaymentReceiveAccessClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenAuthorizingPaymentReceiveAccessBeforeEach(
	t *testing.T,
) *WhenAuthorizingPaymentReceiveAccessTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenAuthorizingPaymentReceiveAccessTestingSuite{
		sut:        services.NewHttpPaymentReceiveAccessClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenAuthorizingPaymentReceiveAccessOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns a payable decision", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalPaymentReceiveAccessWithBodyWithResponse(
					gomock.Any(),
					"application/json",
					gomock.Any(),
				).
				DoAndReturn(func(
					_ context.Context,
					_ string,
					body io.Reader,
					_ ...clients.RequestEditorFn,
				) (*clients.PostInternalPaymentReceiveAccessResponse, error) {
					// The body is hand-marshalled, so its exact keys are the contract.
					var sent map[string]string
					assert.NoError(t, json.NewDecoder(body).Decode(&sent))
					assert.Equal(t, gatePayerAddress, sent["address"])
					assert.Equal(t, "a-header", sent["paymentSignature"])
					return &clients.PostInternalPaymentReceiveAccessResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body: []byte(
							`{"outcome":"payable","payer":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",` +
								`"accepts":[{"network":"eip155:80002","amount":"10000"}],` +
								`"settlement":{"requirements":{"amount":"10000"},` +
								`"facilitatorBaseUrl":"https://facilitator.test","network":"eip155:80002"}}`,
						),
					}, nil
				})
		}

		t.Run("Should decode the verdict and its settlement", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			decision, err := suite.sut.Authorize(context.Background(), gatePayerAddress, "a-header")

			assert.NoError(t, err)
			assert.Equal(t, models.PaymentReceiveAccessOutcomePayable, decision.Outcome)
			assert.Len(t, decision.Accepts, 1)
			if assert.NotNil(t, decision.Settlement) {
				assert.Equal(t, "https://facilitator.test", decision.Settlement.FacilitatorBaseUrl)
				assert.Equal(t, "10000", decision.Settlement.Requirements.Amount)
				assert.Equal(t, "eip155:80002", decision.Settlement.Network)
				assert.Equal(t, gatePayerAddress, decision.Payer)
			}
		})
	})

	t.Run("Given the runtime refuses", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalPaymentReceiveAccessWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalPaymentReceiveAccessResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"outcome":"payment_required","errorCode":"insufficient_amount","accepts":[]}`,
					),
				}, nil)
		}

		t.Run("Should decode the refusal without a settlement", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			decision, err := suite.sut.Authorize(context.Background(), "", "")

			assert.NoError(t, err)
			assert.Equal(t, models.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
			assert.Equal(t, "insufficient_amount", decision.ErrorCode)
			assert.Nil(t, decision.Settlement)
		})
	})

	t.Run("Given the runtime fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingPaymentReceiveAccessTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalPaymentReceiveAccessWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalPaymentReceiveAccessResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"access_evaluation_failed","message":"boom"}`),
				}, nil)
		}

		t.Run("Should map it to an application error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingPaymentReceiveAccessBeforeEach(t)
			initSuite(suite)

			decision, err := suite.sut.Authorize(context.Background(), "", "")

			assert.Nil(t, decision)
			var unavailableError *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableError)
		})
	})
}
