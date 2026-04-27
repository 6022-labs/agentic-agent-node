package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	agent_knowledge_models "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	agent_reasoning_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

func main() {
	dialect := flag.String("dialect", "", "database dialect")
	flag.Parse()

	stmts, err := gormschema.New(*dialect).Load(
		&agent_reasoning_models.MCPConnection{},
		&agent_reasoning_models.LLMConnection{},

		&agent_knowledge_models.ConversationMessage{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
