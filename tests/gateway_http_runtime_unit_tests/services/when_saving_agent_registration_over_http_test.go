package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingAgentRegistrationOverHttpTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSavingAgentRegistrationOverHttpBeforeEach(t *testing.T) *WhenSavingAgentRegistrationOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSavingAgentRegistrationOverHttpTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSavingAgentRegistrationOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime saves the registration draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalAgentRegistrationResponse, error) {
					raw, err := io.ReadAll(reader)
					assert.NoError(t, err)

					var sent models.ExternalSaveAgentRegistrationPayload
					assert.NoError(t, json.Unmarshal(raw, &sent))
					assert.Equal(t, "bob", sent.Name)
					assert.Equal(t, "0xowner", sent.Owner)

					return &clients.PostInternalAgentRegistrationResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"collectionAddress":"0xabc","name":"bob","role":"agent","owner":"0xowner"}`),
					}, nil
				})
		}

		t.Run("Should decode the saved registration payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Role:              "agent",
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "bob", payload.Name)
			}
		})
	})

	t.Run("Given the runtime rejects the registration draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).Return(&clients.PostInternalAgentRegistrationResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
				Body:         []byte(`{"code":"bad","field":"owner","message":"invalid"}`),
			}, nil)
		}

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Role:              "agent",
			})

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})
}
