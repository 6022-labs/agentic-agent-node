package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
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

	sut := use_cases.NewAuthorize(mockAuthorizer)

	return &WhenAuthorizingTestingSuite{
		sut:            sut,
		mockAuthorizer: mockAuthorizer,
	}
}

func TestWhenAuthorizing(t *testing.T) {
	t.Parallel()

	address := "0x3333333333333333333333333333333333333333"

	t.Run("Given the authorizer accepts the address", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingTestingSuite) {
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), common.HexToAddress(address)).Return(true, nil)
		}

		t.Run("Should return Ok with authorized=true", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), &requests.AuthorizeRequest{Address: address})

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.AuthorizeResponse)
			assert.True(t, ok)
			assert.True(t, body.Authorized)
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

			res := suite.sut.Execute(context.Background(), &requests.AuthorizeRequest{Address: address})

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.AuthorizeResponse)
			assert.True(t, ok)
			assert.False(t, body.Authorized)
		})
	})

	t.Run("Given the authorizer fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingTestingSuite) {
			suite.mockAuthorizer.EXPECT().IsAuthorized(gomock.Any(), gomock.Any()).Return(false, errors.New("rpc down"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background(), &requests.AuthorizeRequest{Address: address})

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "rpc down")
		})
	})
}
