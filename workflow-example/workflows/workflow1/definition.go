package workflow1

import (
	"encoding/json"
	"fmt"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

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
				"when": "'$.workflow1.step1.field1' == 'test1'"
			}]
		},
		"step2": {
			"type": "automation",
			"nextSteps": [{
				"name": "step3",
				"when": "'$.workflow1.step1.field2' == 'test2' || '$.workflow1.step1.field3' == 'test2'"
			}]
		},
		"step3": {
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
		"step4": {
			"type": "automation",
			"nextSteps": [{
					"name": "step5",
					"when": "'$.workflow1.step1.field4' < 2"
				},
				{
					"name": "end",
					"when": "'$.workflow1.step1.field4' > 2"
				}
			]
		},
		"step5": {
			"type": "automation",
			"nextSteps": [{
				"name": "step6"
			}]
		},
		"step6": {
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
