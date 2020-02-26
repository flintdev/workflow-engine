package workflow1

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

func TriggerCondition(event workflowFramework.Event) bool {
	return event.Model == "expense" && event.Type == "ADDED"
}
