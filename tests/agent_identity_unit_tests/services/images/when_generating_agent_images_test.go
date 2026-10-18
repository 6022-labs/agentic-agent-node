package images_test

import (
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services/images"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGeneratingAgentImagesTestingSuite struct {
	sut *images.GenerateAgentImages

	mockDefaultImageGenerator *interface_mocks.MockDefaultImageGeneratorInterface
	mockIconImageGenerator    *interface_mocks.MockIconImageGeneratorInterface
}

func WhenGeneratingAgentImagesBeforeEach(t *testing.T) *WhenGeneratingAgentImagesTestingSuite {
	mockController := gomock.NewController(t)

	mockDefaultImageGenerator := interface_mocks.NewMockDefaultImageGeneratorInterface(mockController)
	mockIconImageGenerator := interface_mocks.NewMockIconImageGeneratorInterface(mockController)

	sut := images.NewGenerateAgentImages(mockDefaultImageGenerator, mockIconImageGenerator)

	return &WhenGeneratingAgentImagesTestingSuite{
		sut:                       sut,
		mockDefaultImageGenerator: mockDefaultImageGenerator,
		mockIconImageGenerator:    mockIconImageGenerator,
	}
}

func TestWhenGeneratingAgentImages(t *testing.T) {
	t.Parallel()

	params := identity_interfaces.GenerateImageParams{Name: "my-agent"}

	t.Run("Given the default generator fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGeneratingAgentImagesTestingSuite) {
			suite.mockDefaultImageGenerator.EXPECT().Generate(params).Return(nil, errors.New("default boom"))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingAgentImagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Generate(params)

			assert.ErrorContains(t, err, "default boom")
		})
	})

	t.Run("Given the icon generator fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGeneratingAgentImagesTestingSuite) {
			suite.mockDefaultImageGenerator.EXPECT().Generate(params).Return([]byte("default-svg"), nil)
			suite.mockIconImageGenerator.EXPECT().Generate(params).Return(nil, errors.New("icon boom"))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingAgentImagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Generate(params)

			assert.ErrorContains(t, err, "icon boom")
		})
	})

	t.Run("Given both generators succeed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGeneratingAgentImagesTestingSuite) {
			suite.mockDefaultImageGenerator.EXPECT().Generate(params).Return([]byte("default-svg"), nil)
			suite.mockIconImageGenerator.EXPECT().Generate(params).Return([]byte("icon-png"), nil)
		}

		t.Run("Should key the default and icon bytes the way the collection expects", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingAgentImagesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Generate(params)

			assert.NoError(t, err)
			assert.Equal(t, []byte("default-svg"), result["default"])
			assert.Equal(t, []byte("icon-png"), result["icon"])
		})
	})
}
