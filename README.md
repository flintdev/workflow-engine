# workflow-engine

## File structure

```
|--workflow-example
|  |--main.go
|  |--workflows
|    |--workflow1
|      |--steps
|        |--step1.go
|        |--step2.go
|        |--step3.go
|        |--step4.go
|      |--trigger.go
|      |--definition.json
|      |--index.go
```

## File content

#### main.go

```go
package main

import {
  "encoding/json"
  "fmt"
  "io/ioutil"
  workflow1 "workflows/workflow1"
  workflowFramework "flint/workflow-framework"
}

func main() {
  task1 = workflowFramework.CreateTask()
  task1.registerWorkflowDefinition(workflow1.definition)
  task1.registerSteps(workflow1.steps)
  task1.registerTrigger(workflow1.trigger)
  task1.listen()
}
```

#### step1.go

```go
package steps

func execute(stepData, flowData, handler) {
  // ...
  handler.updateFlowData(newFlowData);
}
```
