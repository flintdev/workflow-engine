package main

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"github.com/flintdev/workflow-engine/workflow-example/workflows"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow1"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow2"
)

func main() {
	app := workflowFramework.CreateApp()
	app.RegisterWorkflow(workflow1.Definition)
	app.RegisterWorkflow(workflow2.Definition)
	app.RegisterConfig(workflows.ParseConfig)
	app.Start()
}
