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

const wfGroup = "flint.flint.com"
const wfVersion = "v1"
const wfResource = "workflows"
const wfNamespace = "default"

func CreateEmptyWorkflowObject(kubeconfig *string, wfObjName string, modelObjName string) error {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flint.flint.com/v1",
			"kind":       "WorkFlow",
			"metadata": map[string]interface{}{
				"name": wfObjName,
				"labels": map[string]interface{}{
					"modelObjName": modelObjName,
					"currentStep":  "init",
				},
			},
			"spec": map[string]interface{}{
				"steps":       []map[string]interface{}{},
				"flowData":    "{}",
				"currentStep": "init",
				"message":     "Init Workflow",
				"status":      "init",
			},
		},
	}
	err := CreateObject(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, obj)
	if err != nil {
		return err
	}
	return nil
}

func GetWorkflowObjectFlowDataValue(kubeconfig *string, objName string, path string) (string, error) {
	path = ParseFlowDataKey(path)
	result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

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
	return m[path], nil
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, wfMessage, "spec", "message"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, status, "spec", "status"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
	err := SetWorkflowObjectStatus(kubeconfig, objName,  "Pending")
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)

		if err != nil {
			return err
		}

		if err := unstructured.SetNestedField(result.Object, currentStep, "spec", "currentStep"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
		if err != nil {
			return err
		}
		if err := unstructured.SetNestedField(result.Object, currentStep, "metadata", "labels", "currentStep"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
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
			"message":   "",
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

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
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
			"status":  status,
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
		if err != nil {
			return err
		}

		flowData, found, err := unstructured.NestedString(result.Object, "spec", "flowData")
		if err != nil || !found || flowData == "" {
			message := fmt.Sprintf("flowData not found or error in spec: %s", err)
			return errors.New(message)
		}

		m, err := ConvertJsonStringToMap(flowData)
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

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func SetWorkflowObjectStepToComplete(kubeconfig *string, objName string, stepName string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Complete")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToRunning(kubeconfig *string, objName string, stepName string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Running")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToPending(kubeconfig *string, objName string, stepName string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Pending")
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflowObjectStepToFailure(kubeconfig *string, objName string, stepName string) error {
	err := setWorkflowObjectStepStatus(kubeconfig, objName, stepName, "Failure")
	if err != nil {
		return err
	}
	return nil
}

func setWorkflowObjectStepStatus(kubeconfig *string, objName string, stepName string, status string) error {
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
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

		if status == "Complete" {
			currentTime := time.Now().UTC().String()
			if err := unstructured.SetNestedField(steps[index].(map[string]interface{}), currentTime, "endAt"); err != nil {
				return err
			}
		}

		if err := unstructured.SetNestedField(result.Object, steps, "spec", "steps"); err != nil {
			return err
		}

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
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
		result, err := GetObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, objName)
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

		res := schema.GroupVersionResource{Group: wfGroup, Version: wfVersion, Resource: wfResource}

		_, err = client.Resource(res).Namespace(wfNamespace).Update(result, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func CheckIfWorkflowIsTriggered(kubeconfig *string, modelObjName string) (bool, error) {
	labelSelector := fmt.Sprintf("modelObjName=%s", modelObjName)
	list, err := ListObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, labelSelector)
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
	labelSelector := fmt.Sprintf("modelObjName=%s, currentStep=%s", modelObjName, currentStep)
	list, err := ListObj(kubeconfig, wfNamespace, wfGroup, wfVersion, wfResource, labelSelector)
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
