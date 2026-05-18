package use_cases_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingBootstrapStatusTestingSuite struct {
	sut *use_cases.GetBootstrapStatus

	mockStateReader *common_mocks.MockBootstrapStateReaderInterface
}

func WhenGettingBootstrapStatusBeforeEach(t *testing.T) *WhenGettingBootstrapStatusTestingSuite {
	mockController := gomock.NewController(t)

	mockStateReader := common_mocks.NewMockBootstrapStateReaderInterface(mockController)

	sut := use_cases.NewGetBootstrapStatus(mockStateReader)

	return &WhenGettingBootstrapStatusTestingSuite{
		sut:             sut,
		mockStateReader: mockStateReader,
	}
}

func TestWhenGettingBootstrapStatus(t *testing.T) {
	t.Parallel()

	t.Run("Given the state reader returns a ready state", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBootstrapStatusTestingSuite) {
			suite.mockStateReader.EXPECT().Get().Return(models.BootstrapState{
				Status:       enums.BootstrapStatusReady,
				AgentAddress: "0xabc",
			})
		}

		t.Run("Should return Ok mirroring the state fields", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBootstrapStatusBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetBootstrapStatusResponse)
			assert.True(t, ok)
			assert.Equal(t, enums.BootstrapStatusReady, body.Status)
			assert.Equal(t, "0xabc", body.AgentAddress)
		})
	})

	t.Run("Given the state reader reports awaiting-vault with a last error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBootstrapStatusTestingSuite) {
			suite.mockStateReader.EXPECT().Get().Return(models.BootstrapState{
				Status:    enums.BootstrapStatusAwaitingVault,
				LastError: "vault sealed",
			})
		}

		t.Run("Should return Ok carrying the LastError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBootstrapStatusBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute()

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.GetBootstrapStatusResponse)
			assert.True(t, ok)
			assert.Equal(t, enums.BootstrapStatusAwaitingVault, body.Status)
			assert.Equal(t, "vault sealed", body.LastError)
		})
	})
}
