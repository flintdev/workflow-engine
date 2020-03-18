package util

import (
	"fmt"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"reflect"
	"strings"
	"time"
)

const wfGroup = "flint.flint.com"
const wfVersion = "v1"
const wfResource = "workflows"
const wfNamespace = "default"

func CreateEmptyWorkflowObject(kubeconfig *string, wfObjName string, modelObjName string) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flint.flint.com/v1",
			"kind":       "WorkFlow",
			"metadata": map[string]interface{}{
				"name": wfObjName,
				"labels": map[string]interface{}{
					"modelObjName": modelObjName,
					"currentStep": "init",
				},
			},
			"spec": map[string]interface{}{
				"steps":       []map[string]interface{}{},
				"flowData":    "{}",
				"currentStep": "init",
				"message": "Init Workflow",
				"status": "init",
			},
		},
	}
	CreateObject(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, obj)
}

func GetWorkflowObjectFlowDataValue(kubeconfig *string, objName string, path string) string {
	path = ParseFlowDataKey(path)
	result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

	flowData, found, err := unstructured.NestedString(result.Object, "spec", "flowData")
	if err != nil || !found || flowData == "" {
		panic(fmt.Errorf("flowData not found or error in spec: %v", err))
	}

	m := ConvertJsonStringToMap(flowData)
	return m[path]
}

func GetWorkflowObjectCurrentStep(kubeconfig *string, objName string) string {
	result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

	currentStep, found, err := unstructured.NestedString(result.Object, "spec", "currentStep")
	if err != nil || !found || currentStep == "" {
		panic(fmt.Errorf("currentStep not found or error in spec: %v", err))
	}
	return currentStep
}

func SetWorkflowObjectCurrentStep(kubeconfig *string, objName string, currentStep string) {
	fmt.Println(1)
	fmt.Println(kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	fmt.Println(2)

	if err != nil {
		panic(err)
	}
	fmt.Println(3)

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

		fmt.Println(result)

		if err := unstructured.SetNestedField(result.Object, currentStep, "spec", "currentStep"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, updateErr := client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
}

func SetWorkflowObjectCurrentStepLabel(kubeconfig *string, objName string, currentStep string) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
		if err := unstructured.SetNestedField(result.Object, currentStep, "metadata", "labels", "currentStep"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, updateErr := client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
}

func SetStepToWorkflowObject(kubeconfig *string, stepName string, objName string) {
	status := "Running"
	currentTime := time.Now().UTC().String()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			panic(fmt.Errorf("steps not found or error in spec: %v", err))
		}
		tempStep := map[string]interface{}{
			"name":    stepName,
			"startAt": currentTime,
			"endAt":   "",
			"status":  status,
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, updateErr := client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
}

func SetWorkflowObjectFlowData(kubeconfig *string, objName string, path string, value string) {
	path = ParseFlowDataKey(path)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

		flowData, found, err := unstructured.NestedString(result.Object, "spec", "flowData")
		if err != nil || !found || flowData == "" {
			panic(fmt.Errorf("flowData not found or error in spec: %v", err))
		}

		m := ConvertJsonStringToMap(flowData)
		m[path] = value
		jsonString := ConvertMapToJsonString(m)

		if err := unstructured.SetNestedField(result.Object, jsonString, "spec", "flowData"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, updateErr := client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
}

func SetWorkflowObjectStepToComplete(kubeconfig *string, objName string, stepName string) {
	setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Complete")
}

func SetWorkflowObjectStepToRunning(kubeconfig *string, objName string, stepName string) {
	setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Running")
}

func SetWorkflowObjectStepToPending(kubeconfig *string, objName string, stepName string) {
	setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Pending")
}

func SetWorkflowObjectStepToFailure(kubeconfig *string, objName string, stepName string) {
	setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Failure")
}

func setWorkflowObjectStepStatus(kubeconfig *string, objName string, stepName string, status string) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	var index int

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			panic(fmt.Errorf("steps not found or error in spec: %v", err))
		}
		for i, step := range steps {
			getIndex := false
			v := reflect.ValueOf(step)
			if v.Kind() == reflect.Map {
				for _, key := range v.MapKeys() {
					stepValue := v.MapIndex(key).Interface()
					if stepValue == stepName {
						index = i
						getIndex = true
						break
					}
				}
			}
			if getIndex {
				break
			}
		}
		if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), status, "status"); err != nil {
			panic(err)
		}

		if status == "Complete" {
			currentTime := time.Now().UTC().String()
			if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), currentTime, "endAt"); err != nil {
				panic(err)
			}
		}

		if err := unstructured.SetNestedField(result.Object, steps, "spec", "steps"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, updateErr := client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}

}

func CheckIfWorkflowIsTriggered(kubeconfig *string, modelObjName string) bool{
	fmt.Println("checking if workflow is triggered by Model Event")
	labelSelector := fmt.Sprintf("modelObjName=%s",  modelObjName)
	list := ListObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, labelSelector)
	fmt.Println(len(list.Items))
	if len(list.Items) > 0 {
		return true
	} else {
		return false
	}
}

func GetPendingWorkflowList(kubeconfig *string, modelObjName string, currentStep string) *unstructured.UnstructuredList{
	labelSelector := fmt.Sprintf("modelObjName=%s, currentStep=%s",  modelObjName, currentStep)
	list := ListObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, labelSelector)
	return list
}

func GenerateWorkflowObjName() string {
	uuidWithHyphen := uuid.New()
	u := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	name := "workflow-" + u
	return name
}

func ParseFlowDataKey(path string) string {
	s := strings.Split(path, ".")
	if s[0] == "$" {
		s = s[1:]
	}
	return strings.Join(s[:], ".")
}
