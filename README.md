# workflow-engine

## File structure

```
|--workflow-example
|  |--main.go
|  |--config.json
|  |--workflows
|    |--workflow1
|      |--steps
         |--step1
|          |--step1.go
|        |--step2
|          |--step2.go
|        |--step3
|          |--step3.go
|          |--step4
|        |--step4.go
|      |--trigger.go
|      |--definition.json
|      |--init.go
```

## File content

#### main.go

```go
package main

import (
	workflowFramework "workflow-engine/engine"
	"workflow-engine/workflow-example/workflows"
	workflow1 "workflow-engine/workflow-example/workflows/workflow1"
)

func main() {
	app := workflowFramework.CreateApp()
	app.RegisterWorkflow(workflow1.Definition, workflow1.Steps, workflow1.Trigger)
	app.RegisterConfig(workflows.ParseConfig)
	app.Start()
}

```

#### step1.go

```go
package step1

import (
	"fmt"
	"time"
	"workflow-engine/handler"
)

func Execute(handler handler.Handler) {
	fmt.Println("running step1")
	path := "step1.field1.field2"
	value := "test1"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}


```

#### trigger.go

```go
package workflow1

import (
	workflowFramework "workflow-engine/engine"
)

func TriggerCondition(event workflowFramework.Event) bool {
	return event.Model == "expense" && event.Type == "ADDED"
}

```
