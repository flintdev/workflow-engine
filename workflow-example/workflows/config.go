package workflows

import (
	"encoding/json"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"go.uber.org/zap"
)

func ParseConfig() workflowFramework.Config {
	var c workflowFramework.Config
	config := `{
	"gvr": {
		"approval": {
			"group": "flint.flint.com",
			"version": "v1",
			"resource": "approvals"
		},
		"expense": {
			"group": "flint.flint.com",
			"version": "v1",
			"resource": "expenses"
		}
	}
}`
	err := json.Unmarshal([]byte(config), &c)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err != nil {
		logger.Error(err.Error())
	}
	return c

}
