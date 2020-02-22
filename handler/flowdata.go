package handler

import "workflow-engine/util"

func SetFlowData(kubeconfig *string, objName string, path string, value string) {
	util.SetWorkflowObjectFlowData(kubeconfig, objName, path, value)
}

func GetFlowDataByPath(kubeconfig *string, objName string, path string) string {
	r := util.GetWorkflowObjectFlowDataValue(kubeconfig, objName, path)
	return r
}
