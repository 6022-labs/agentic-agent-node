package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenPublishingEnsRuntimeRecordsTestingSuite struct {
	sut *use_cases.PublishEnsRuntimeRecords

	mockPublisher *identity_mocks.MockEnsRuntimeRecordsPublisherInterface
}

func WhenPublishingEnsRuntimeRecordsBeforeEach(t *testing.T) *WhenPublishingEnsRuntimeRecordsTestingSuite {
	mockController := gomock.NewController(t)

	mockPublisher := identity_mocks.NewMockEnsRuntimeRecordsPublisherInterface(mockController)

	sut := use_cases.NewPublishEnsRuntimeRecords(mockPublisher)

	return &WhenPublishingEnsRuntimeRecordsTestingSuite{
		sut:           sut,
		mockPublisher: mockPublisher,
	}
}

func TestWhenPublishingEnsRuntimeRecords(t *testing.T) {
	t.Parallel()

	t.Run("Given the publisher succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingEnsRuntimeRecordsTestingSuite) {
			suite.mockPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(nil)
		}

		t.Run("Should return NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingEnsRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the publisher reports ErrIdentityNotFound", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingEnsRuntimeRecordsTestingSuite) {
			suite.mockPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(common_interfaces.ErrIdentityNotFound)
		}

		t.Run("Should return Conflict with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingEnsRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "identity_not_found", conflictErr.Code)
		})
	})

	t.Run("Given the publisher returns an unexpected error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingEnsRuntimeRecordsTestingSuite) {
			suite.mockPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(errors.New("ens registry down"))
		}

		t.Run("Should return InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingEnsRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "ens registry down")
		})
	})
}
