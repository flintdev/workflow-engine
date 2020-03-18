package workflow1

import (
	"encoding/json"
	"fmt"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

// Example of event.Type == MODIFIED

func ParseDefinition() workflowFramework.Workflow {
	var w workflowFramework.Workflow
	definition := `{
	"name": "workflow1",
	"startAt": "step1",
	"trigger": {
		"model": "expense",
		"eventType": "MODIFIED",
		"when": "'spec.switch' == 'true'"
	},
	"steps": {
		"step1": {
			"type": "automation",
			"nextSteps": [{
					"name": "step2",
					"condition": {
						"key": "$.workflow1.step1.field1.field2",
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
			"type": "manual",
			"trigger": {
				"model": "expense",
				"eventType": "MODIFIED",
				"when": "'spec.approval' == 'true'"
			},
			"nextSteps": [{
				"name": "step4"
			}]
		},
		"step3": {
			"type": "automation",
			"nextSteps": [{
				"name": "step4"
			}]
		},
		"step4": {
			"type": "automation",
			"nextSteps": [{
				"name": "end"
			}]
		},
		"end": {
			"nextSteps": []
		}
	}
}`
	err := json.Unmarshal([]byte(definition), &w)

	if err != nil {
		fmt.Println(err.Error())
	}
	return w

}
