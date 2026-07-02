package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	agent_identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	agent_knowledge_models "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	agent_payment_models "github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
)

func main() {
	dialect := flag.String("dialect", "", "database dialect")
	flag.Parse()

	stmts, err := gormschema.New(*dialect).Load(
		&mcpConnectionWithCrossDomainFK{},
		&llmConnectionWithCrossDomainFK{},

		&conversationWithCrossDomainFK{},
		&agent_knowledge_models.ConversationMessage{},

		&agent_identity_models.Agent{},
		&agent_identity_models.AgentRegistrationState{},
		&agent_identity_models.AgentAddress{},

		&paymentPolicyWithCrossDomainFK{},
		&agent_payment_models.PaymentRule{},
		&agent_payment_models.PaymentRuleX402Config{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
