package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/metrics_mocks/interfaces_mocks"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenPublishingEnsRuntimeRecordsTestingSuite struct {
	sut *use_cases.PublishEnsRuntimeRecords

	mockEnsPublisher *identity_mocks.MockEnsPublisherInterface
}

func WhenPublishingEnsRuntimeRecordsBeforeEach(t *testing.T) *WhenPublishingEnsRuntimeRecordsTestingSuite {
	mockController := gomock.NewController(t)

	mockEnsPublisher := identity_mocks.NewMockEnsPublisherInterface(mockController)

	mockEnsPublicationMetrics := metrics_mocks.NewMockEnsPublicationMetricsInterface(mockController)
	mockEnsPublicationMetrics.EXPECT().RecordEnsPublishedSucceeded(gomock.Any()).AnyTimes()
	mockEnsPublicationMetrics.EXPECT().RecordEnsPublishedFailed(gomock.Any()).AnyTimes()
	mockEnsPublicationMetrics.EXPECT().RecordEnsPublishedNotMinted(gomock.Any()).AnyTimes()

	sut := use_cases.NewPublishEnsRuntimeRecords(mockEnsPublisher, mockEnsPublicationMetrics)

	return &WhenPublishingEnsRuntimeRecordsTestingSuite{
		sut:              sut,
		mockEnsPublisher: mockEnsPublisher,
	}
}

func TestWhenPublishingEnsRuntimeRecords(t *testing.T) {
	t.Parallel()

	t.Run("Given the publisher reports ErrIdentityNotFound", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingEnsRuntimeRecordsTestingSuite) {
			suite.mockEnsPublisher.EXPECT().
				PublishRuntimeRecords(gomock.Any()).
				Return(common_interfaces.ErrIdentityNotFound)
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
			suite.mockEnsPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(errors.New("ens registry down"))
		}

		t.Run("Should return an InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingEnsRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "ens registry down")
		})
	})

	t.Run("Given the publisher succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenPublishingEnsRuntimeRecordsTestingSuite) {
			suite.mockEnsPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(nil)
		}

		t.Run("Should return NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenPublishingEnsRuntimeRecordsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})
}
