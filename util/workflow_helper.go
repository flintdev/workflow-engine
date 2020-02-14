package util

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

func CreateSampleWorkflowObject(kubeconfig *string) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flint.flint.com/v1",
			"kind":       "WorkFlow",
			"metadata": map[string]interface{}{
				"name": "workflow-sample-01",
			},
			"spec": map[string]interface{}{
				"steps": []map[string]interface{}{
					{
						"name":   "step1",
						"output": "test",
						"status": "complete",
					},
				},
			},
		},
	}
	CreateObject(kubeconfig, "default", "flint.flint.com", "v1", "workflows", obj)
}

func UpdateSampleObj(kubeconfig *string, namespace string, group string, version string, resource string, step string, output string, status string) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result := GetObj(kubeconfig, namespace, group, "v1", resource)
		steps, found, err := unstructured.NestedSlice(result.Object, "spec", "steps")
		if err != nil || !found || steps == nil {
			panic(fmt.Errorf("steps not found or error in spec: %v", err))
		}
		tempStep := map[string]interface{}{

			"name":   step,
			"output": output,
			"status": "complete",
		}
		newSteps := append(steps, tempStep)

		if err := unstructured.SetNestedField(result.Object, newSteps, "spec", "steps"); err != nil {
			panic(err)
		}

		res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

		_, updateErr := client.Resource(res).Namespace(namespace).Update(result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}

}
