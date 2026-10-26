package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSigningTypedDataTestingSuite struct {
	sut *use_cases.SignTypedData

	mockTypedDataSigner *interfaces_mocks.MockTypedDataSignerInterface
}

func WhenSigningTypedDataBeforeEach(t *testing.T) *WhenSigningTypedDataTestingSuite {
	mockController := gomock.NewController(t)

	mockTypedDataSigner := interfaces_mocks.NewMockTypedDataSignerInterface(mockController)

	return &WhenSigningTypedDataTestingSuite{
		sut:                 use_cases.NewSignTypedData(mockTypedDataSigner),
		mockTypedDataSigner: mockTypedDataSigner,
	}
}

func TestWhenSigningTypedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	typedData := &apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}},
			"Mail":         {{Name: "contents", Type: "string"}},
		},
		PrimaryType: "Mail",
		Domain:      apitypes.TypedDataDomain{Name: "Test"},
		Message:     apitypes.TypedDataMessage{"contents": "hello"},
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{}

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			suite.mockTypedDataSigner.EXPECT().SignTypedData(gomock.Any(), gomock.Any()).Times(0)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the signer fails", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSigner.EXPECT().
				SignTypedData(ctx, typedData).
				Return("", errors.New("node wallet not bootstrapped"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "typed_data_signing_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given the wallet address is unavailable", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSigner.EXPECT().SignTypedData(ctx, typedData).Return("0xdead", nil)
			suite.mockTypedDataSigner.EXPECT().
				WalletAddress(ctx).
				Return(common.Address{}, errors.New("node wallet not bootstrapped"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "wallet_address_unavailable", internalErr.Code)
			}
		})
	})

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SignTypedDataRequest{TypedData: typedData}
		address := common_tests.RandomEthAddress()
		initSuite := func(suite *WhenSigningTypedDataTestingSuite) {
			suite.mockTypedDataSigner.EXPECT().SignTypedData(ctx, typedData).Return("0xdead", nil)
			suite.mockTypedDataSigner.EXPECT().WalletAddress(ctx).Return(address, nil)
		}

		t.Run("Should return the signature and the signing address", func(t *testing.T) {
			t.Parallel()

			suite := WhenSigningTypedDataBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			if assert.NotNil(t, response) {
				assert.Equal(t, "0xdead", response.Signature)
				assert.Equal(t, address.Hex(), response.Address)
			}
		})
	})
}
