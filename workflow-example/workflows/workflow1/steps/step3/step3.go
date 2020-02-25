package step3

import (
	"fmt"
	"time"
	"workflow-engine/handler"
)

func Execute(kubeconfig *string, objName string) {
	fmt.Println("running step3")
	path := "step2.field1.field2"
	value := "test4"
	handler.SetFlowData(kubeconfig, objName, path, value)
	time.Sleep(5 * time.Second)
}
