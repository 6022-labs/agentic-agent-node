package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentIdentityTestingSuite struct {
	sut *use_cases.GetAgentIdentity

	mockResolver *common_mocks.MockAgentIdentityResolverInterface
}

func WhenGettingAgentIdentityBeforeEach(t *testing.T) *WhenGettingAgentIdentityTestingSuite {
	mockController := gomock.NewController(t)

	mockResolver := common_mocks.NewMockAgentIdentityResolverInterface(mockController)

	sut := use_cases.NewGetAgentIdentity(mockResolver)

	return &WhenGettingAgentIdentityTestingSuite{
		sut:          sut,
		mockResolver: mockResolver,
	}
}

func TestWhenGettingAgentIdentity(t *testing.T) {
	t.Parallel()

	t.Run("Given the resolver returns an identity", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(&models.AgentIdentity{
				CollectionAddress: "0xcoll",
				TokenId:           42,
				Name:              "Alice",
			}, nil)
		}

		t.Run("Should return Ok with the identity on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetAgentIdentityResponse)
			assert.True(t, ok)
			assert.Equal(t, "0xcoll", body.CollectionAddress)
			assert.Equal(t, uint64(42), body.TokenId)
			assert.Equal(t, "Alice", body.Name)
		})
	})

	t.Run("Given the resolver returns ErrIdentityNotFound", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(nil, common_interfaces.ErrIdentityNotFound)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNotFound, res.Outcome)
			assert.Equal(t, "identity_not_found", res.Code)
		})
	})

	t.Run("Given the resolver returns an unexpected error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentIdentityTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(nil, errors.New("rpc down"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentIdentityBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "rpc down")
		})
	})
}
