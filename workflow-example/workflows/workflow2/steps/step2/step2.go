package step2

import (
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"time"
)

func Execute(handler handler.Handler) {
	fmt.Println("running workflow2 step2")
	path := "$.workflow2.step2.field1.field2"
	value := "test2"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}
