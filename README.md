# workflow-engine

## File structure

```
|--workflow-example
|  |--main.go
|  |--config.go
|  |--go.mod
|  |--workflows
|    |--workflow1
|      |--definition.go
|      |--init.go
```

## File content

#### main.go

```go
package main

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"github.com/flintdev/workflow-engine/workflow-example/workflows"
	"workflow-example/workflows/workflow1"
	"workflow-example/workflows/workflow2"
)

func main() {
	app := workflowFramework.CreateApp()
	app.RegisterWorkflow(workflow1.Definition)
	app.RegisterWorkflow(workflow2.Definition)
	app.RegisterConfig(workflows.ParseConfig)
	app.Start()
}

```

#### definition.go

```go
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
					"condition": {
						"key": "$.workflow1.step1.field1",
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


```

#### init.go

```go
package workflow1

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
)

func Definition() workflowFramework.Workflow {
	w := ParseDefinition()
	return w
}


```
