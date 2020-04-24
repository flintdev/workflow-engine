package workflow1

import (
	"encoding/json"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"go.uber.org/zap"
)

func ParseDefinition() workflowFramework.Workflow {
	var w workflowFramework.Workflow
	definition := `{
	"name": "workflow1",
	"startAt": ["step1"],
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
				"when": "\"$.workflow1.step1.field1\" == \"test1\""
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
				"eventType": "MODIFIED"
			},
			"nextSteps": [{
					"name": "step4",
					"when": "'spec.approval' == 'true'"
				},
				{
					"name": "end",
					"when": "'spec.approval' == 'false'"
				}
			]
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
					"name": "step6",
					"when": "'$.workflow1.step1.field5' < '2020-03-01'"
				},
				{
					"name": "step7",
					"when": "'$.workflow1.step1.field5' >= '2020-03-01'"
				}
			]
		},
		"step6": {
			"type": "automation",
			"nextSteps": [{
				"name": "end"
			}]
		},
		"step7": {
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

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err != nil {
		logger.Error(err.Error())
	}

	return w

}
