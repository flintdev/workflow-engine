package step1

import (
	"fmt"
	"time"
	"workflow-engine/handler"
)

func Execute(kubeconfig *string, objName string) {
	fmt.Println("running step1")
	path := "step1.field1.field2"
	value := "test1"
	handler.SetFlowData(kubeconfig, objName, path, value)
	time.Sleep(5 * time.Second)
}
