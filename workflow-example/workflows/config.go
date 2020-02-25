package workflows

import (
	"encoding/json"
	"fmt"
	workflowFramework "workflow-engine/engine"
)

func ParseConfig() workflowFramework.Config {
	var c workflowFramework.Config
	config := `{
  "gvr": {
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
