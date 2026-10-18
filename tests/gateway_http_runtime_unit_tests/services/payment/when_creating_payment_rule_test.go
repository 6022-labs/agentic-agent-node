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

type WhenCreatingPaymentRuleTestingSuite struct {
	sut *services.HttpPaymentRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingPaymentRuleBeforeEach(t *testing.T) *WhenCreatingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingPaymentRuleTestingSuite{
		sut:        services.NewHttpPaymentRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentRule{Name: "Public price"})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigPaymentPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"asset","message":"no ERC-20 token answers"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(context.Background(), &models.ExternalPaymentRule{})

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "asset", validationErr.Field)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigPaymentPolicyRulesResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentRule{Name: "Public price"})

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigPaymentPolicyRulesResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalPaymentRule{Name: "Public price"})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, body io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigPaymentPolicyRulesResponse, error) {
					var sent models.ExternalPaymentRule
					assert.NoError(t, json.NewDecoder(body).Decode(&sent))
					assert.Equal(t, "Public price", sent.Name)
					return &clients.PostInternalConfigPaymentPolicyRulesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"r1","name":"Public price"}`),
					}, nil
				})
		}

		t.Run("Should forward and decode the created rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(context.Background(), &models.ExternalPaymentRule{Name: "Public price"})

			assert.NoError(t, err)
			assert.Equal(t, "r1", rule.Id)
		})
	})
}
