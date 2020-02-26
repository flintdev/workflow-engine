package workflow1

import (
	"encoding/json"
	"fmt"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

func ParseDefinition() workflowFramework.Workflow {
	var w workflowFramework.Workflow
	definition := `{
  "name": "test01",
  "startAt": "step1",
  "steps": {
    "step1": {
      "nextSteps": [
        {
          "name": "step2",
          "condition": {
            "key": "step1.field1.field2",
            "value": "test1",
            "operator": "="
          }
        },
        {
          "name": "step3",
          "condition": {
            "key": "step1.result",
            "value": "Failure",
            "operator": "="
          }
        }
      ]
    },
    "step2": {
      "nextSteps": [
        {
          "name": "step4"
        }
      ]
    },
    "step3": {
      "nextSteps": [
        {
          "name": "step4"
        }
      ]
    },
    "step4": {
      "nextSteps": [
        {
          "name": "end"
        }
      ]
    },
    "end": {
      "nextSteps": [
      ]
    }
  }
}`
	err := json.Unmarshal([]byte(definition), &w)

	if err != nil {
		fmt.Println(err.Error())
	}
	return w

}
