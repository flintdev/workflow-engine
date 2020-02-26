package step4

import (
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"time"
)

func Execute(handler handler.Handler) {
	fmt.Println("running step4")
	path := "$.step2.field1.field2"
	value := "test4"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}
