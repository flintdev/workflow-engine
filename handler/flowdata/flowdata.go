package flowdata

import "github.com/flintdev/workflow-engine/util"

type FlowData struct {
	Kubeconfig *string
	WFObjName  string
}

func (fd *FlowData) Set(path string, value string) error {
	kubeconfig := fd.Kubeconfig
	objName := fd.WFObjName
	err := util.SetWorkflowObjectFlowData(kubeconfig, objName, path, value)
	if err != nil {
		return err
	}
	return nil
}

func (fd *FlowData) Get(path string) (string, error) {
	kubeconfig := fd.Kubeconfig
	objName := fd.WFObjName
	r, err := util.GetWorkflowObjectFlowDataValue(kubeconfig, objName, path)
	if err != nil {
		return "", err
	}
	return r, err
}
