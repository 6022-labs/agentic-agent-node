package use_cases_test

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/identity_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingAgentRegistrationTestingSuite struct {
	sut *use_cases.SaveAgentRegistration

	mockClient *interfaces_mocks.MockAgentRegistrationClientInterface
}

func WhenSavingAgentRegistrationBeforeEach(t *testing.T) *WhenSavingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)
	mockClient := interfaces_mocks.NewMockAgentRegistrationClientInterface(mockController)
	sut := use_cases.NewSaveAgentRegistration(mockClient)
	return &WhenSavingAgentRegistrationTestingSuite{sut: sut, mockClient: mockClient}
}

func TestWhenSavingAgentRegistration(t *testing.T) {
	t.Parallel()

	owner := strings.ToLower(common_tests.RandomEthAddress().Hex())
	ownerCtx := helpers.ContextWithCallerId(context.Background(), owner)
	validRequest := func() *requests.SaveAgentRegistrationRequest {
		return &requests.SaveAgentRegistrationRequest{
			ChainId:           "80002",
			CollectionAddress: "0x2222222222222222222222222222222222222222",
			Name:              "Alice",
			AttributesRaw:     `{"role":"assistant"}`,
			Image:             &multipart.FileHeader{Filename: "avatar.png", Size: 128},
		}
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			_, err := suite.sut.Execute(ownerCtx, &requests.SaveAgentRegistrationRequest{Name: "Alice"})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), validRequest())

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_owner", unauthorizedErr.Code)
		})
	})

	t.Run("Given a caller id that is not a wallet address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			ctx := helpers.ContextWithCallerId(context.Background(), "apikey:7f3c")

			_, err := suite.sut.Execute(ctx, validRequest())

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "invalid_owner", unauthorizedErr.Code)
		})
	})

	t.Run("Given the client returns a domain error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().Save(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("draft_not_found", "no draft"))
		}

		t.Run("Should propagate it unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ownerCtx, validRequest())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "draft_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime accepts the draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingAgentRegistrationTestingSuite) *models.ExternalSaveAgentRegistrationPayload {
			forwarded := &models.ExternalSaveAgentRegistrationPayload{}
			suite.mockClient.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, body *models.ExternalSaveAgentRegistrationPayload) (*models.ExternalAgentRegistrationPayload, error) {
					*forwarded = *body
					return &models.ExternalAgentRegistrationPayload{Name: body.Name}, nil
				})
			return forwarded
		}

		t.Run("Should forward the authenticated caller as the checksummed owner", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			forwarded := initSuite(suite)

			_, err := suite.sut.Execute(ownerCtx, validRequest())

			assert.NoError(t, err)
			assert.Equal(t, common.HexToAddress(owner).Hex(), forwarded.Owner)
			assert.Equal(t, "Alice", forwarded.Name)
		})

		t.Run("Should return the saved registration", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ownerCtx, validRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Alice", resp.Name)
		})
	})
}
