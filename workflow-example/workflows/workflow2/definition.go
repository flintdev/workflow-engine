package workflow2

import (
	"encoding/json"
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"go.uber.org/zap"
)

func ParseDefinition() workflowFramework.Workflow {
	var w workflowFramework.Workflow
	definition := `{
	"name": "workflow2",
	"startAt": "step1",
	"trigger": {
		"model": "approval",
		"eventType": "ADDED"
	},
	"steps": {
		"step1": {
			"type": "automation",
			"nextSteps": [{
					"name": "step2"
				},
				{
					"name": "step3"
				}
			]
		},
		"step2": {
			"type": "automation",
			"nextSteps": [{
				"name": "step4"
			}]
		},
		"step3": {
			"type": "automation",
			"nextSteps": [{
				"name": "step6"
			}]
		},
		"step4": {
			"type": "automation",
			"nextSteps": [{
				"name": "step5"
			}]
		},
		"step5": {
			"type": "automation",
			"nextSteps": [{
				"name": "step7"
			}]
		},
		"step6": {
			"type": "manual",
			"trigger": {
				"model": "approval",
				"eventType": "MODIFIED",
				"when": "'spec.approval' == 'true'"
			},
			"nextSteps": [{
				"name": "step7"
			}]
		},
		"step7": {
			"type": "hub",
			"inputs": ["step5", "step6"],
			"condition": "all_success",
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
