package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	gateway_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWhenVerifyingAuth(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x3333333333333333333333333333333333333333")
	req := &requests.VerifyAuthRequest{
		Address:   address.Hex(),
		Signature: "0xsig",
		Nonce:     "nonce-xyz",
	}

	t.Run("Given the nonce is missing or expired", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(false)

		uc := use_cases.NewVerifyAuth(nonceStore, nil, nil, nil)

		t.Run("Should return ErrNonceNotFound", func(t *testing.T) {
			_, err := uc.Execute(context.Background(), req)
			assert.ErrorIs(t, err, use_cases.ErrNonceNotFound)
		})
	})

	t.Run("Given the recovered signer differs from the claimed address", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(true)
		verifier := gateway_mocks.NewMockSignatureVerifierInterface(ctrl)
		verifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(common.HexToAddress("0xdeadbeef"), nil)

		uc := use_cases.NewVerifyAuth(nonceStore, verifier, nil, nil)

		t.Run("Should return ErrSignatureMismatch", func(t *testing.T) {
			_, err := uc.Execute(context.Background(), req)
			assert.ErrorIs(t, err, use_cases.ErrSignatureMismatch)
		})
	})

	t.Run("Given the signature verifier fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(true)
		verifier := gateway_mocks.NewMockSignatureVerifierInterface(ctrl)
		verifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(common.Address{}, errors.New("malformed"))

		uc := use_cases.NewVerifyAuth(nonceStore, verifier, nil, nil)

		t.Run("Should propagate the verifier error", func(t *testing.T) {
			_, err := uc.Execute(context.Background(), req)
			assert.ErrorContains(t, err, "malformed")
		})
	})

	t.Run("Given the authorizer rejects the address", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(true)
		verifier := gateway_mocks.NewMockSignatureVerifierInterface(ctrl)
		verifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
		authorizer := common_mocks.NewMockAuthorizerInterface(ctrl)
		authorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, nil)

		uc := use_cases.NewVerifyAuth(nonceStore, verifier, authorizer, nil)

		t.Run("Should return ErrAddressNotAuthorized", func(t *testing.T) {
			_, err := uc.Execute(context.Background(), req)
			assert.ErrorIs(t, err, use_cases.ErrAddressNotAuthorized)
		})
	})

	t.Run("Given the authorizer errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(true)
		verifier := gateway_mocks.NewMockSignatureVerifierInterface(ctrl)
		verifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
		authorizer := common_mocks.NewMockAuthorizerInterface(ctrl)
		authorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(false, errors.New("rpc down"))

		uc := use_cases.NewVerifyAuth(nonceStore, verifier, authorizer, nil)

		t.Run("Should propagate the authorizer error", func(t *testing.T) {
			_, err := uc.Execute(context.Background(), req)
			assert.ErrorContains(t, err, "rpc down")
		})
	})

	t.Run("Given a valid nonce, signature, and authorized address", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedExp := time.Now().Add(30 * time.Minute)
		nonceStore := gateway_mocks.NewMockNonceStoreInterface(ctrl)
		nonceStore.EXPECT().Consume("nonce-xyz").Return(true)
		verifier := gateway_mocks.NewMockSignatureVerifierInterface(ctrl)
		verifier.EXPECT().Recover("nonce-xyz", "0xsig").Return(address, nil)
		authorizer := common_mocks.NewMockAuthorizerInterface(ctrl)
		authorizer.EXPECT().IsAuthorized(gomock.Any(), address).Return(true, nil)
		issuer := gateway_mocks.NewMockJwtIssuerInterface(ctrl)
		issuer.EXPECT().Issue(address).Return("the-token", expectedExp, nil)

		uc := use_cases.NewVerifyAuth(nonceStore, verifier, authorizer, issuer)

		t.Run("Should return a JWT response", func(t *testing.T) {
			resp, err := uc.Execute(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, "the-token", resp.Token)
			assert.Equal(t, expectedExp, resp.ExpiresAt)
		})
	})
}
