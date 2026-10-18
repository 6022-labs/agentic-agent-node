package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/identity"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSigningAgentCardTestingSuite struct {
	sut *services.HttpAgentCardSignatureClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
	sentBody   []byte
}

func WhenSigningAgentCardBeforeEach(t *testing.T) *WhenSigningAgentCardTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSigningAgentCardTestingSuite{
		sut:        services.NewHttpAgentCardSignatureClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSigningAgentCard(t *testing.T) {
	t.Parallel()

	const card = `{"name":"hermes","protocol":"6022"}`

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalIdentityCardSignatureWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Sign(context.Background(), []byte(card))

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalIdentityCardSignatureWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalIdentityCardSignatureResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"card","message":"agent card must not carry a signatures member"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Sign(context.Background(), []byte(card))

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalIdentityCardSignatureWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalIdentityCardSignatureResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Sign(context.Background(), []byte(card))

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime signs the card", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSigningAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalIdentityCardSignatureWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, body io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalIdentityCardSignatureResponse, error) {
					sent, err := io.ReadAll(body)
					assert.NoError(t, err)
					suite.sentBody = sent

					return &clients.PostInternalIdentityCardSignatureResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"protected":"cHJvdGVjdGVk","signature":"c2ln"}`),
					}, nil
				})
		}

		t.Run("Should decode the detached JWS", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningAgentCardBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Sign(context.Background(), []byte(card))

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "cHJvdGVjdGVk", payload.Protected)
				assert.Equal(t, "c2ln", payload.Signature)
			}
		})

		t.Run(
			"Should forward the card bytes verbatim so the runtime signs what the gateway serves",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenSigningAgentCardBeforeEach(t)
				initSuite(suite)

				_, err := suite.sut.Sign(context.Background(), []byte(card))
				assert.NoError(t, err)

				var sent struct {
					Card json.RawMessage `json:"card"`
				}
				assert.NoError(t, json.Unmarshal(suite.sentBody, &sent))
				assert.JSONEq(t, card, string(sent.Card))
			},
		)
	})
}
