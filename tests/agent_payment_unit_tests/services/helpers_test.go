package services_test

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
)

func generateSaveRequest(config string, price string) *requests.SavePaymentReceiveRuleRequest {
	return &requests.SavePaymentReceiveRuleRequest{
		Price:  price,
		Config: json.RawMessage(config),
	}
}
