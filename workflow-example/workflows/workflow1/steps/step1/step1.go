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
