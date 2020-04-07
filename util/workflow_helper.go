package util

import (
	"errors"
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

const WFGroup = "flint.flint.com"
const WFVersion = "v1"
const WFResource = "workflows"
const WFNamespace = "default"

func CreateEmptyWorkflowObject(kubeconfig *string, wfObjName string, modelObjName string) error {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flint.flint.com/v1",
			"kind":       "WorkFlow",
			"metadata": map[string]interface{}{
				"name": wfObjName,
				"labels": map[string]interface{}{
					"modelObjName": modelObjName,
				},
			},
			"spec": map[string]interface{}{
				"steps":       []map[string]interface{}{},
				"flowData":    "{}",
				"currentStep": "init",
				"status":      "init",
				"message":     "",
			},
		},
	}
	err := CreateObject(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, obj)
	if err != nil {
		return err
	}
	return nil
}

func GetWorkflowObjectStatus(kubeconfig *string, objName string) (string, error) {
	result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)

	if err != nil {
		return "", err
	}

	status, found, err := unstructured.NestedString(result.Object, "spec", "status")
	if err != nil || !found || status == "" {
		message := fmt.Sprintf("status not found or error in spec: %s", err)
		return "", errors.New(message)
	}

	return status, nil
}

func GetWorkflowObjectFlowDataValue(kubeconfig *string, objName string, path string) (interface{}, error) {
	path = ParseFlowDataKey(path)
	result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)

	if err != nil {
		return "", err
	}

	flowData, found, err := unstructured.NestedString(result.Object, "spec", "flowData")
	if err != nil || !found || flowData == "" {
		message := fmt.Sprintf("flowData not found or error in spec: %s", err)
		return "", errors.New(message)
	}
	m, err := ConvertJsonStringToMap(flowData)
	if err != nil {
		return "", err
	}
	if value, exist := m[path]; exist {
		return value, nil
	} else {
		message := fmt.Sprintf("key %s is not found in flow data", path)
		return nil, errors.New(message)
	}
}

func SetWorkflowObjectMessage(kubeconfig *string, objName string, wfMessage string) error {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, wfMessage, "spec", "message"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectStatus(kubeconfig *string, objName string, status string) error {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, status, "spec", "status"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectToComplete(kubeconfig *string, objName string) error {
	err := SetWorkflowObjectStatus(kubeconfig, objName, "Complete")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectToRunning(kubeconfig *string, objName string) error {
	err := SetWorkflowObjectStatus(kubeconfig, objName, "Running")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectToPending(kubeconfig *string, objName string) error {
	err := SetWorkflowObjectStatus(kubeconfig, objName, "Pending")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectToFailure(kubeconfig *string, objName string, wfMessage string) error {
	err := SetWorkflowObjectStatus(kubeconfig, objName, "Failure")
	if err != nil {
		return err
	}
	err = SetWorkflowObjectMessage(kubeconfig, objName, wfMessage)
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectCurrentStep(kubeconfig *string, objName string, currentStep string) error {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, currentStep, "spec", "currentStep"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectCurrentStepLabel(kubeconfig *string, objName string, currentStep string) error {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		if err := unstructured.SetNestedField(result.Object, currentStep, "metadata", "labels", "currentStep"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectPendingStepLabel(kubeconfig *string, objName string, stepName string) error {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		if err := unstructured.SetNestedField(result.Object, "Pending", "metadata", "labels", stepName); err != nil {
			return err
		}
		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func RemoveWorkflowObjectPendingStepLabel(kubeconfig *string, objName string, stepName string) error {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		unstructured.RemoveNestedField(result.Object, "metadata", "labels", stepName)

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectStep(kubeconfig *string, objName string, stepName string) error {
	status := "Running"
	currentTime := time.Now().UTC().String()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			message := fmt.Sprintf("steps not found or error in spec: %s", err)
			return errors.New(message)
		}
		tempStep := map[string]interface{}{
			"name":    stepName,
			"startAt": currentTime,
			"endAt":   "",
			"message": "",
			"status":  status,
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, stepName, "spec", "currentStep"); err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, stepName, "metadata", "labels", "currentStep"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetStepToWorkflowObject(kubeconfig *string, stepName string, objName string) error {
	status := "Running"
	currentTime := time.Now().UTC().String()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			message := fmt.Sprintf("steps not found or error in spec: %s", err)
			return errors.New(message)
		}
		tempStep := map[string]interface{}{
			"name":    stepName,
			"startAt": currentTime,
			"endAt":   "",
			"message": "",
			"status":  status,
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetPendingStepToWorkflowObject(kubeconfig *string, stepName string, objName string) error {
	status := "Pending"
	currentTime := time.Now().UTC().String()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			message := fmt.Sprintf("steps not found or error in spec: %s", err)
			return errors.New(message)
		}
		tempStep := map[string]interface{}{
			"name":    stepName,
			"startAt": currentTime,
			"endAt":   "",
			"message": "",
			"status":  status,
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectFlowData(kubeconfig *string, objName string, path string, value string) error {
	path = ParseFlowDataKey(path)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}

		flowData, found, err := unstructured.NestedString(result.Object, "spec", "flowData")
		if err != nil || !found || flowData == "" {
			message := fmt.Sprintf("flowData not found or error in spec: %s", err)
			return errors.New(message)
		}

		m, err := ConvertJsonStringToStringMap(flowData)
		if err != nil {
			return err
		}
		m[path] = value
		jsonString, err := ConvertMapToJsonString(m)
		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, jsonString, "spec", "flowData"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectStepToComplete(kubeconfig *string, objName string, stepName string, message string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Complete", message)
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToRunning(kubeconfig *string, objName string, stepName string, message string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Running", message)
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToPending(kubeconfig *string, objName string, stepName string, message string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Pending", message)
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToFailure(kubeconfig *string, objName string, stepName string, message string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Failure", message)
	if err != nil {
		return err
	}
	return nil
}

func setWorkflowObjectStepStatus(kubeconfig *string, objName string, stepName string, status string, message string) error {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	var index int

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			message := fmt.Sprintf("steps not found or error in spec: %s", err)
			return errors.New(message)
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
			return err
		}

		if message != "" {
			if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), message, "message"); err != nil {
				return err
			}
		}

		if status == "Complete" || status == "Failure" {
			currentTime := time.Now().UTC().String()
			if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), currentTime, "endAt"); err != nil {
				return err
			}
		}

		if err := unstructured.SetNestedField(result.Object, steps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil

}

func SetWorkflowObjectFailedStep(kubeconfig *string, objName string, stepName string, stepMessage string) error {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	var index int

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
		if err != nil {
			return err
		}
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			message := fmt.Sprintf("steps not found or error in spec: %s", err)
			return errors.New(message)
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
		if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), stepMessage, "message"); err != nil {
			return err
		}
		if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), "Failure", "status"); err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, steps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: WFGroup, Version: WFVersion, Resource: WFResource}

		_, err = client.Resource(res).Namespace(WFNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func CheckAllStepStatus(kubeconfig *string, objName string) (string, string, error) {
	// return enum: allCompleteSuccess, hasRunning, allCompleteHasFailure, hasPending

	var failureSteps []string

	result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
	if err != nil {
		return "", "", err
	}
	steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
	if err != nil || !found || steps == nil {
		message := fmt.Sprintf("steps not found or error in spec: %s", err)
		return "", "", errors.New(message)
	}
	hasFailure := false
	for _, step := range steps {
		stepName := ""
		status := ""
		v := reflect.ValueOf(step)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				if key.Interface() == "status" {
					status = v.MapIndex(key).Interface().(string)
				}
				if key.Interface() == "name" {
					stepName = v.MapIndex(key).Interface().(string)
				}
			}
		}
		switch status {
		case "Running":
			return "hasRunning", "", nil
		case "Pending":
			return "hasPending", "", nil
		case "Complete":
		case "Failure":
			hasFailure = true
			failureSteps = append(failureSteps, stepName)
		}
	}
	if hasFailure {
		message := fmt.Sprintf("Failed on steps: %s", strings.Join(failureSteps[:], ","))
		return "allCompleteHasFailure", message, nil
	}
	return "allCompleteSuccess", "", nil
}

func CheckAllStepStatusByList(kubeconfig *string, objName string, stepsList []string) (string, error) {
	// return enum: all_success, all_failed
	var statusList []string

	result, err := GetObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, objName)
	if err != nil {
		return "", err
	}
	steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
	if err != nil || !found || steps == nil {
		message := fmt.Sprintf("steps not found or error in spec: %s", err)
		return "", errors.New(message)
	}
	for _, step := range steps {
		stepName := ""
		status := ""
		v := reflect.ValueOf(step)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				if key.Interface() == "status" {
					status = v.MapIndex(key).Interface().(string)
				}
				if key.Interface() == "name" {
					stepName = v.MapIndex(key).Interface().(string)
				}
			}
		}
		for _, name := range stepsList {
			if name == stepName {
				statusList = append(statusList, status)
			}
		}
	}

	for _, status := range statusList {
		if status == "Complete" {
			continue
		} else {
			return "no", nil
		}
	}
	return "all_success", nil
}

func CheckIfWorkflowIsTriggered(kubeconfig *string, modelObjName string) (bool, error) {
	labelSelector := fmt.Sprintf("modelObjName=%s", modelObjName)
	list, err := ListObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, labelSelector)
	if err != nil {
		return false, err
	}
	if len(list.Items) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func GetPendingWorkflowList(kubeconfig *string, modelObjName string, currentStep string) (*unstructured.UnstructuredList, error) {
	var errorReturn *unstructured.UnstructuredList
	labelSelector := fmt.Sprintf("modelObjName=%s, %s=%s", modelObjName, currentStep, "Pending")
	list, err := ListObj(kubeconfig, WFNamespace, WFGroup, WFVersion, WFResource, labelSelector)
	if err != nil {
		return errorReturn, err
	}
	return list, nil
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
