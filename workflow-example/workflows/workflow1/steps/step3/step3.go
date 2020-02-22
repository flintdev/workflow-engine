package step3

import (
	"fmt"
	"time"
)

func Execute(kubeconfig *string, objName string) {
	fmt.Println("running step3")
	time.Sleep(5 * time.Second)
}
