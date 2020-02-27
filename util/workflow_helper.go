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

func CreateEmptyWorkflowObject(kubeconfig *string, objName string) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flint.flint.com/v1",
			"kind":       "WorkFlow",
			"metadata": map[string]interface{}{
				"name": objName,
			},
			"spec": map[string]interface{}{
				"steps":    []map[string]interface{}{},
				"flowData": "{}",
			},
		},
	}
	CreateObject(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, obj)
}

func AddStepToWorkflowObject(kubeconfig *string, stepName string, objName string) {
	status := "running"
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

func SetWorkflowObjectStepToComplete(kubeconfig *string, objName string, stepName string) {
	setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Complete")
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
