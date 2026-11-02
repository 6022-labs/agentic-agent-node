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

type WhenCreatingPaymentReceiveRuleTestingSuite struct {
	sut *services.HttpPaymentReceiveRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenCreatingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingPaymentReceiveRuleTestingSuite{
		sut:        services.NewHttpPaymentReceiveRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentReceiveRule{Name: "Public price"})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigPaymentReceivePolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"asset","message":"no ERC-20 token answers"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(context.Background(), &models.ExternalPaymentReceiveRule{})

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "asset", validationErr.Field)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigPaymentReceivePolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body: []byte(
						`{"code":"asset_already_has_rule","message":"another rule already prices this token"}`,
					),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentReceiveRule{Name: "Public price"})

			var target *apperrors.ConflictError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "asset_already_has_rule", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigPaymentReceivePolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentReceiveRule{Name: "Public price"})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, body io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigPaymentReceivePolicyRulesResponse, error) {
					var sent models.ExternalPaymentReceiveRule
					assert.NoError(t, json.NewDecoder(body).Decode(&sent))
					assert.Equal(t, "Public price", sent.Name)
					assert.Equal(t, []models.ExternalPaymentReceiveRulePrice{
						{Price: "1000", AddressPatterns: []string{".*"}},
					}, sent.Prices)
					return &clients.PostInternalConfigPaymentReceivePolicyRulesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"r1","name":"Public price"}`),
					}, nil
				})
		}

		t.Run("Should forward and decode the created rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(
				context.Background(),
				&models.ExternalPaymentReceiveRule{
					Name:   "Public price",
					Prices: []models.ExternalPaymentReceiveRulePrice{{Price: "1000", AddressPatterns: []string{".*"}}},
				},
			)

			assert.NoError(t, err)
			assert.Equal(t, "r1", rule.Id)
		})
	})
}
