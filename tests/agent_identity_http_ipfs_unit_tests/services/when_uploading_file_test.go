package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_http_ipfs_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenUploadingFileTestingSuite struct {
	sut *services.IPFSFileStorage

	mockClient *clients_mocks.MockIPFSClientInterface
}

func WhenUploadingFileBeforeEach(t *testing.T) *WhenUploadingFileTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockIPFSClientInterface(mockController)

	return &WhenUploadingFileTestingSuite{
		sut:        services.NewIPFSFileStorage(zap.NewNop(), mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUploadingFile(t *testing.T) {
	t.Parallel()

	t.Run("Given the client adds the file", func(t *testing.T) {
		t.Parallel()

		// RoutingProvide runs in a background goroutine — synchronise on it so
		// the call is recorded before the controller is finished.
		initSuite := func(suite *WhenUploadingFileTestingSuite) chan string {
			provided := make(chan string, 1)
			suite.mockClient.EXPECT().Add([]byte("image-bytes"), "agent.png", gomock.Any()).DoAndReturn(
				func(_ []byte, _ string, params *models.ExternalAddQueryRequest) (*models.ExternalAddResponse, error) {
					if assert.NotNil(t, params) && assert.NotNil(t, params.Pin) {
						assert.True(t, *params.Pin)
					}
					return &models.ExternalAddResponse{Hash: "QmHashValue"}, nil
				})
			suite.mockClient.EXPECT().RoutingProvideWithoutResponse(gomock.Any()).DoAndReturn(
				func(params *models.ExternalRoutingProvideRequest) error {
					provided <- params.Arg
					return nil
				})
			return provided
		}

		t.Run("Should pin the upload, return the hash, and announce it to the DHT", func(t *testing.T) {
			t.Parallel()

			suite := WhenUploadingFileBeforeEach(t)
			provided := initSuite(suite)

			result, err := suite.sut.UploadFile([]byte("image-bytes"), "agent.png")

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "QmHashValue", *result)
			}

			select {
			case arg := <-provided:
				assert.Equal(t, "QmHashValue", arg)
			case <-time.After(2 * time.Second):
				t.Fatal("expected the cid to be announced to the DHT")
			}
		})
	})

	t.Run("Given the client fails to add the file", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUploadingFileTestingSuite) {
			suite.mockClient.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("ipfs down"))
		}

		t.Run("Should return the error without announcing anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenUploadingFileBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.UploadFile([]byte("x"), "f.png")

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}
