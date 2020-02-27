package workflows

import (
	"encoding/json"
	"fmt"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
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

	if err != nil {
		fmt.Println(err.Error())
	}
	return c

}
