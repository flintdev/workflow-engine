package util

import (
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type GVR struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

func GetKubeConfig() *string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return kubeconfig
}

func CreateObject(kubeconfig *string, namespace string, group string, version string, resource string, obj *unstructured.Unstructured) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	result, err := client.Resource(res).Namespace(namespace).Create(obj, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created Workflow Object", result)
}

func GetObj(kubeconfig *string, namespace string, group string, version string, resource string, objName string) *unstructured.Unstructured {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	result, getErr := client.Resource(res).Namespace(namespace).Get(objName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("failed to get latest version of Expense: %v", getErr))
	}

	return result
}

func ListObj(kubeconfig *string, namespace string, group string, version string, resource string, labelSelector string) (*unstructured.UnstructuredList, error) {
	var errReturn *unstructured.UnstructuredList
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return errReturn, err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return errReturn, err
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	list, err := client.Resource(res).Namespace(namespace).List(listOptions)
	if err != nil {
		return errReturn, err
	}

	return list, nil
}

func WatchObject(kubeconfig *string, namespace string, group string, version string, resource string) <-chan watch.Event {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	watcher, err := client.Resource(res).Namespace(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	ch := watcher.ResultChan()

	return ch
}
