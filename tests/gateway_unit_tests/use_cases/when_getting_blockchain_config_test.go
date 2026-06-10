package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingBlockchainConfigTestingSuite struct {
	sut *use_cases.GetBlockchainConfig

	mockClient *interfaces_mocks.MockBlockchainConfigClientInterface
}

func WhenGettingBlockchainConfigBeforeEach(t *testing.T) *WhenGettingBlockchainConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockBlockchainConfigClientInterface(mockController)

	sut := use_cases.NewGetBlockchainConfig(mockClient)

	return &WhenGettingBlockchainConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenGettingBlockchainConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&models.ExternalBlockchainConfigPayload{
				ChainId:                   80002,
				CollectionsManagerAddress: "0xmanager",
				CollectionCreatorAddress:  "0xcreator",
				Token6022Address:          "0xtoken",
				EnsRegistryAddress:        "0xens",
			}, nil)
		}

		t.Run("Should map every field", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, uint64(80002), resp.ChainId)
			assert.Equal(t, "0xmanager", resp.CollectionsManagerAddress)
			assert.Equal(t, "0xcreator", resp.CollectionCreatorAddress)
			assert.Equal(t, "0xtoken", resp.Token6022Address)
			assert.Equal(t, "0xens", resp.EnsRegistryAddress)
		})
	})

	t.Run("Given a prior successful fetch", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&models.ExternalBlockchainConfigPayload{
				ChainId:                   80002,
				CollectionsManagerAddress: "0xmanager",
				CollectionCreatorAddress:  "0xcreator",
				Token6022Address:          "0xtoken",
				EnsRegistryAddress:        "0xens",
			}, nil).Times(1)
		}

		t.Run("Should serve the second call from cache", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)
			initSuite(suite)

			first, err := suite.sut.Execute(context.Background())
			assert.NoError(t, err)
			second, err := suite.sut.Execute(context.Background())
			assert.NoError(t, err)

			assert.Same(t, first, second)
		})
	})

	t.Run("Given the runtime is down on the first call", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigTestingSuite) {
			gomock.InOrder(
				suite.mockClient.EXPECT().Get(gomock.Any()).
					Return(nil, errors.NewUnavailableError("runtime_unavailable", "down")),
				suite.mockClient.EXPECT().Get(gomock.Any()).Return(&models.ExternalBlockchainConfigPayload{
					ChainId:                   80002,
					CollectionsManagerAddress: "0xmanager",
					CollectionCreatorAddress:  "0xcreator",
					Token6022Address:          "0xtoken",
					EnsRegistryAddress:        "0xens",
				}, nil),
			)
		}

		t.Run("Should not cache the error and succeed on retry", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())
			var target *errors.UnavailableError
			assert.ErrorAs(t, err, &target)

			resp, err := suite.sut.Execute(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, uint64(80002), resp.ChainId)
		})
	})
}
