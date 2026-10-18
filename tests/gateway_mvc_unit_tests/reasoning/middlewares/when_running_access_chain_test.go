package middlewares_test

import (
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
	payment_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	access_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mvc_mocks/reasoning_mocks/middlewares_mocks/interfaces_mocks"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenRunningAccessChainTestingSuite struct {
	mockPaymentRuleClient *payment_mocks.MockPaymentRuleClientInterface
	linkA                 *access_mocks.MockAccessMethodInterface
	linkB                 *access_mocks.MockAccessMethodInterface
	downstreamRan         bool
}

func WhenRunningAccessChainBeforeEach(t *testing.T) *WhenRunningAccessChainTestingSuite {
	mockController := gomock.NewController(t)
	return &WhenRunningAccessChainTestingSuite{
		mockPaymentRuleClient: payment_mocks.NewMockPaymentRuleClientInterface(mockController),
		linkA:                 access_mocks.NewMockAccessMethodInterface(mockController),
		linkB:                 access_mocks.NewMockAccessMethodInterface(mockController),
	}
}

// build wires the gate over the given links and exposes /responses.
func (s *WhenRunningAccessChainTestingSuite) build(links ...middleware_interfaces.AccessMethodInterface) *fiber.App {
	sut := middlewares.NewResponseAccessGateMiddleware(middlewares.NewResponseAccessGateMiddlewareParams{
		PaymentRuleClient: s.mockPaymentRuleClient,
		AccessMethods:     links,
	})
	app := fiber.New()
	app.Post("/responses", sut.Handle, func(ctx fiber.Ctx) error {
		s.downstreamRan = true
		return ctx.JSON(fiber.Map{"ok": true})
	})
	return app
}

func accessChainBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func TestWhenRunningAccessChain(t *testing.T) {
	t.Parallel()

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return 502 without consulting any link", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, errors.New("runtime down")).Times(1)
			suite.linkA.EXPECT().Priority().Return(10).AnyTimes()

			resp, err := suite.build(suite.linkA).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
			assert.Contains(t, accessChainBody(t, resp), "payment_rules_unavailable")
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should serve the request through the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet")).
				Times(1)
			suite.linkA.EXPECT().Priority().Return(10).AnyTimes()
			suite.linkA.EXPECT().TryHandle(gomock.Any(), gomock.Len(0)).Return(true, nil).Times(1)

			resp, err := suite.build(suite.linkA).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	})

	t.Run("Given a lower-priority link is registered after a higher one", func(t *testing.T) {
		t.Parallel()

		t.Run("Should run links in priority order and stop at the first to handle", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, nil).Times(1)
			// linkA registered first but is the lower-priority (later) link.
			suite.linkA.EXPECT().Priority().Return(100).AnyTimes()
			suite.linkB.EXPECT().Priority().Return(10).AnyTimes()
			// linkB runs first and handles; linkA must never be consulted.
			suite.linkB.EXPECT().TryHandle(gomock.Any(), gomock.Any()).Return(true, nil).Times(1)

			resp, err := suite.build(suite.linkA, suite.linkB).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.False(t, suite.downstreamRan)
		})
	})

	t.Run("Given the first link defers", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fall through to the next link in order", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, nil).Times(1)
			suite.linkA.EXPECT().Priority().Return(10).AnyTimes()
			suite.linkB.EXPECT().Priority().Return(100).AnyTimes()
			gomock.InOrder(
				suite.linkA.EXPECT().TryHandle(gomock.Any(), gomock.Any()).Return(false, nil),
				suite.linkB.EXPECT().TryHandle(gomock.Any(), gomock.Any()).Return(true, nil),
			)

			resp, err := suite.build(suite.linkA, suite.linkB).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	})

	t.Run("Given the handling link calls next", func(t *testing.T) {
		t.Parallel()

		t.Run("Should run the downstream handler", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, nil).Times(1)
			suite.linkA.EXPECT().Priority().Return(10).AnyTimes()
			suite.linkA.EXPECT().
				TryHandle(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx fiber.Ctx, _ []models.ExternalPaymentRule) (bool, error) {
					return true, ctx.Next()
				}).
				Times(1)

			resp, err := suite.build(suite.linkA).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.downstreamRan)
		})
	})

	t.Run("Given no link handles the request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return 502 — the terminal link is misconfigured", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, nil).Times(1)
			suite.linkA.EXPECT().Priority().Return(10).AnyTimes()
			suite.linkA.EXPECT().TryHandle(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

			resp, err := suite.build(suite.linkA).Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
			assert.Contains(t, accessChainBody(t, resp), "access_methods_unavailable")
			assert.False(t, suite.downstreamRan)
		})
	})

	t.Run("Given the chain is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return 502", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningAccessChainBeforeEach(t)
			suite.mockPaymentRuleClient.EXPECT().List(gomock.Any()).Return(nil, nil).Times(1)

			resp, err := suite.build().Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
			assert.Contains(t, accessChainBody(t, resp), "access_methods_unavailable")
		})
	})
}
