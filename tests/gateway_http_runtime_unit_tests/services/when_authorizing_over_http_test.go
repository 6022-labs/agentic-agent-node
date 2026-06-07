package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingOverHttpTestingSuite struct {
	sut *services.HttpAuthorizer

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenAuthorizingOverHttpBeforeEach(t *testing.T) *WhenAuthorizingOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenAuthorizingOverHttpTestingSuite{
		sut:        services.NewHttpAuthorizer(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenAuthorizingOverHttp(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x1111111111111111111111111111111111111111")

	t.Run("Given the runtime authorizes the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAuthorizeWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalAuthorizeResponse, error) {
					var sent models.ExternalAuthorizePayload
					raw, _ := io.ReadAll(reader)
					_ = json.Unmarshal(raw, &sent)
					assert.Equal(t, address.Hex(), sent.Address)
					return &clients.PostInternalAuthorizeResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"authorized":true}`),
					}, nil
				})
		}

		t.Run("Should send the address and return true", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.NoError(t, err)
			assert.True(t, authorized)
		})
	})

	t.Run("Given the runtime denies the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAuthorizeWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalAuthorizeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"authorized":false}`),
				}, nil)
		}

		t.Run("Should return false", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAuthorizeWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
