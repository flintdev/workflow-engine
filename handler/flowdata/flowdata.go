package flowdata

import "github.com/flintdev/workflow-engine/util"

type FlowData struct {
	Kubeconfig *string
	WFObjName  string
}

func (fd *FlowData) Set(path string, value string) {
	kubeconfig := fd.Kubeconfig
	objName := fd.WFObjName
	util.SetWorkflowObjectFlowData(kubeconfig, objName, path, value)
}

func (fd *FlowData) Get(path string) string {
	kubeconfig := fd.Kubeconfig
	objName := fd.WFObjName
	r := util.GetWorkflowObjectFlowDataValue(kubeconfig, objName, path)
	return r
}
