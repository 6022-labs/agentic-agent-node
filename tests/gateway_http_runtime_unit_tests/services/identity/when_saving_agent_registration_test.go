package services_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/identity"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingAgentRegistrationTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSavingAgentRegistrationBeforeEach(t *testing.T) *WhenSavingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSavingAgentRegistrationTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSavingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Attributes:        map[string]string{"role": "agent"},
				Image:             httpTestFileHeader(t, []byte("img"), "image/png"),
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"code":"bad","field":"owner","message":"invalid"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Attributes:        map[string]string{"role": "agent"},
				Image:             httpTestFileHeader(t, []byte("img"), "image/png"),
			})

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Attributes:        map[string]string{"role": "agent"},
				Image:             httpTestFileHeader(t, []byte("img"), "image/png"),
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime saves the registration draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(
					func(_ context.Context, contentType string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalAgentRegistrationResponse, error) {
						assert.Contains(t, contentType, "multipart/form-data")

						buffered, err := io.ReadAll(reader)
						assert.NoError(t, err)
						assert.Equal(t, "bob", readMultipartField(t, contentType, bytes.NewReader(buffered), "name"))
						assert.Equal(
							t,
							"0xowner",
							readMultipartField(t, contentType, bytes.NewReader(buffered), "owner"),
						)
						assert.Equal(
							t,
							"base.png",
							readMultipartField(t, contentType, bytes.NewReader(buffered), "image"),
						)

						return &clients.PostInternalAgentRegistrationResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"collectionAddress":"0xabc","name":"bob","role":"agent","owner":"0xowner"}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the saved registration payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Save(context.Background(), &models.ExternalSaveAgentRegistrationPayload{
				CollectionAddress: "0xabc",
				Owner:             "0xowner",
				Name:              "bob",
				Attributes:        map[string]string{"role": "agent"},
				Image:             httpTestFileHeader(t, []byte("img"), "image/png"),
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "bob", payload.Name)
			}
		})
	})
}
