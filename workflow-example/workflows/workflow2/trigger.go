package workflow2

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

func TriggerCondition(event workflowFramework.Event) bool {
	return event.Model == "approval" && event.Type == "ADDED"
}
