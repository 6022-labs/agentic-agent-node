package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/metrics_mocks/interfaces_mocks"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingTestingSuite struct {
	sut *use_cases.Authorize

	mockAuthorizer *common_mocks.MockAuthorizerInterface
}

func WhenAuthorizingBeforeEach(t *testing.T) *WhenAuthorizingTestingSuite {
	mockController := gomock.NewController(t)

	mockAuthorizer := common_mocks.NewMockAuthorizerInterface(mockController)

	mockAuthorizationMetrics := metrics_mocks.NewMockAuthorizationMetricsInterface(mockController)
	mockAuthorizationMetrics.EXPECT().RecordAuthorizationAuthorized(gomock.Any()).AnyTimes()
	mockAuthorizationMetrics.EXPECT().RecordAuthorizationDenied(gomock.Any()).AnyTimes()
	mockAuthorizationMetrics.EXPECT().RecordAuthorizationFailed(gomock.Any()).AnyTimes()

	sut := use_cases.NewAuthorize(mockAuthorizer, mockAuthorizationMetrics)

	return &WhenAuthorizingTestingSuite{
		sut:            sut,
		mockAuthorizer: mockAuthorizer,
	}
}

func TestWhenAuthorizing(t *testing.T) {
	t.Parallel()

	address := "0x3333333333333333333333333333333333333333"

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.AuthorizeRequest{
					WalletAddressRequest: requests.WalletAddressRequest{Address: "not-an-address"},
				},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the authorizer fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingTestingSuite) {
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), gomock.Any()).Return(false, errors.New("rpc down"))
		}

		t.Run("Should return an InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.AuthorizeRequest{WalletAddressRequest: requests.WalletAddressRequest{Address: address}},
			)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.ErrorContains(t, internalErr.Cause, "rpc down")
		})
	})

	t.Run("Given the authorizer rejects the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingTestingSuite) {
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), common.HexToAddress(address)).Return(false, nil)
		}

		t.Run("Should return Ok with authorized=false", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.AuthorizeRequest{WalletAddressRequest: requests.WalletAddressRequest{Address: address}},
			)

			assert.NoError(t, err)
			assert.False(t, resp.Authorized)
		})
	})

	t.Run("Given the authorizer accepts the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingTestingSuite) {
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), common.HexToAddress(address)).Return(true, nil)
		}

		t.Run("Should return Ok with authorized=true", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.AuthorizeRequest{WalletAddressRequest: requests.WalletAddressRequest{Address: address}},
			)

			assert.NoError(t, err)
			assert.True(t, resp.Authorized)
		})
	})
}
