package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingBootstrapStatusTestingSuite struct {
	sut *use_cases.GetBootstrapStatus

	mockStateReader      *common_mocks.MockBootstrapStateReaderInterface
	mockIdentityResolver *common_mocks.MockAgentIdentityResolverInterface
}

func WhenGettingBootstrapStatusBeforeEach(t *testing.T) *WhenGettingBootstrapStatusTestingSuite {
	mockController := gomock.NewController(t)

	mockStateReader := common_mocks.NewMockBootstrapStateReaderInterface(mockController)
	mockIdentityResolver := common_mocks.NewMockAgentIdentityResolverInterface(mockController)

	sut := use_cases.NewGetBootstrapStatus(mockStateReader, mockIdentityResolver)

	return &WhenGettingBootstrapStatusTestingSuite{
		sut:                  sut,
		mockStateReader:      mockStateReader,
		mockIdentityResolver: mockIdentityResolver,
	}
}

func TestWhenGettingBootstrapStatus(t *testing.T) {
	t.Parallel()

	t.Run("Given the wallet is still awaiting vault", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBootstrapStatusTestingSuite) {
			suite.mockStateReader.EXPECT().Get().Return(models.BootstrapState{
				Status: enums.BootstrapStatusAwaitingVault,
			})
		}

		t.Run("Should return Ok without attempting identity resolution", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBootstrapStatusBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetBootstrapStatusResponse)
			assert.True(t, ok)
			assert.Equal(t, enums.BootstrapStatusAwaitingVault, body.Status)
			assert.Nil(t, body.Identity)
		})
	})

	t.Run("Given the wallet is ready and the identity resolves", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBootstrapStatusTestingSuite) {
			suite.mockStateReader.EXPECT().Get().Return(models.BootstrapState{
				Status:       enums.BootstrapStatusReady,
				AgentAddress: "0xabc",
			})
			suite.mockIdentityResolver.EXPECT().Resolve(gomock.Any()).Return(&models.AgentIdentity{
				CollectionAddress: "0xcoll",
				TokenId:           7,
				Name:              "alice",
			}, nil)
		}

		t.Run("Should return Ok with a resolved identity snapshot", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBootstrapStatusBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetBootstrapStatusResponse)
			assert.True(t, ok)
			if assert.NotNil(t, body.Identity) {
				assert.True(t, body.Identity.Resolved)
				assert.Equal(t, "0xcoll", body.Identity.CollectionAddress)
				assert.Equal(t, uint64(7), body.Identity.TokenId)
				assert.Equal(t, "alice", body.Identity.Name)
			}
		})
	})

	t.Run("Given the wallet is ready but identity resolution fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBootstrapStatusTestingSuite) {
			suite.mockStateReader.EXPECT().Get().Return(models.BootstrapState{
				Status:       enums.BootstrapStatusReady,
				AgentAddress: "0xabc",
			})
			suite.mockIdentityResolver.EXPECT().Resolve(gomock.Any()).Return(nil, errors.New("nft not minted"))
		}

		t.Run("Should return Ok with an unresolved identity carrying the error inline", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBootstrapStatusBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetBootstrapStatusResponse)
			assert.True(t, ok)
			if assert.NotNil(t, body.Identity) {
				assert.False(t, body.Identity.Resolved)
				assert.Contains(t, body.Identity.Error, "nft not minted")
			}
		})
	})
}
