package step2

import (
	"fmt"
	"time"
	"workflow-engine/handler"
)

func Execute(handler handler.Handler) {
	fmt.Println("running step2")
	path := "step2.field1.field2"
	value := "test2"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}
