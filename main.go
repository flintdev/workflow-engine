package main

import (
	workflowFramework "workflow-engine/engine"
	workflow1 "workflow-engine/workflow-example/workflows/workflow1"
)

func main() {
	task1 := workflowFramework.CreateTask()
	task1.RegisterWorkflowDefinition(workflow1.Definition)
	task1.RegisterSteps(workflow1.Steps)
	task1.RegisterTrigger(workflow1.Trigger)
	task1.Listen()
}
