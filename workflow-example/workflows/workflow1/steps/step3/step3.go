package step3

import (
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"time"
)

func Execute(handler handler.Handler) {
	fmt.Println("running workflow1  step3")
	path := "$.workflow1.step3.field1.field3"
	value := "test3"
	handler.FlowData.Set(path, value)
	time.Sleep(5 * time.Second)
}
