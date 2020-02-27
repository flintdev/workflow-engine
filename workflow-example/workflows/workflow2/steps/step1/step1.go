package step1

import (
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"time"
)

func Execute(handler handler.Handler) {
	fmt.Println("running workflow2 step1")
	path := "$.workflow2.step1.field1.field2"
	value := "test1"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}
