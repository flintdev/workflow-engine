package workflow2

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

func Definition() workflowFramework.Workflow {
	w := ParseDefinition()
	return w
}
