package mvc_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenWritingAnErrorBodyTestingSuite struct {
	sut  *mvc.ErrorWriter
	logs *observer.ObservedLogs
}

func WhenWritingAnErrorBodyBeforeEach(t *testing.T) *WhenWritingAnErrorBodyTestingSuite {
	core, logs := observer.New(zapcore.DebugLevel)
	return &WhenWritingAnErrorBodyTestingSuite{
		sut:  mvc.NewErrorWriter(zap.New(core)),
		logs: logs,
	}
}

func (s *WhenWritingAnErrorBodyTestingSuite) write(t *testing.T, status int, code string, body any) (int, string) {
	app := fiber.New()
	app.Get("/pay", func(ctx fiber.Ctx) error {
		return s.sut.WriteBody(ctx, status, code, body)
	})

	resp, testErr := app.Test(httptest.NewRequest(http.MethodGet, "/pay", nil))
	assert.NoError(t, testErr)
	raw, readErr := io.ReadAll(resp.Body)
	assert.NoError(t, readErr)
	assert.NoError(t, resp.Body.Close())
	return resp.StatusCode, string(raw)
}

func TestWhenWritingAnErrorBody(t *testing.T) {
	t.Parallel()

	t.Run("Given a protocol-shaped payment required body", func(t *testing.T) {
		t.Parallel()

		body := fiber.Map{"x402Version": 2, "error": "insufficient_funds"}

		t.Run("Should respond with the status and the body unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAnErrorBodyBeforeEach(t)

			status, raw := suite.write(t, http.StatusPaymentRequired, "insufficient_funds", body)

			assert.Equal(t, http.StatusPaymentRequired, status)
			assert.JSONEq(t, `{"x402Version":2,"error":"insufficient_funds"}`, raw)
		})

		t.Run("Should log a warning with the code and response", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAnErrorBodyBeforeEach(t)

			suite.write(t, http.StatusPaymentRequired, "insufficient_funds", body)

			logs := suite.logs.All()
			assert.Len(t, logs, 1)
			assert.Equal(t, zapcore.WarnLevel, logs[0].Level)
			fields := logs[0].ContextMap()
			assert.Equal(t, "insufficient_funds", fields["code"])
			assert.Equal(t, int64(http.StatusPaymentRequired), fields["status"])
			assert.JSONEq(t, `{"x402Version":2,"error":"insufficient_funds"}`, fields["response"].(string))
		})
	})
}
