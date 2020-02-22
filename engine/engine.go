package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
	"workflow-engine/util"
)

type FlowData map[string]string

type StepCondition struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Operator string `json:"operator"`
}

type NextStep struct {
	Name      string        `json:"name"`
	Condition StepCondition `json:"condition"`
}

type NextSteps struct {
	NextSteps []NextStep `json:"nextSteps"`
}

type Workflow struct {
	Name    string               `json:"name"`
	StartAt string               `json:"startAt"`
	Steps   map[string]NextSteps `json:"steps"`
}

type WorkflowInstance struct {
	Workflow Workflow
	StepFunc map[string]func(kubeconfig *string, objName string)
	Trigger  func(event Event) bool
}

type App struct {
	WorkflowInstances []WorkflowInstance
	ModelGVRMap       map[string]GVR
}

type Event struct {
	Type  string
	Model string
}

type GVR struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

type Config struct {
	GVRMap map[string]GVR `json:"gvr"`
}

func CreateWorkflowInstance() WorkflowInstance {
	var wi WorkflowInstance
	return wi
}

func CreateApp() App {
	var app App
	return app
}

func ParseDefinition(filePath string) Workflow {
	var w Workflow
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		panic(fmt.Errorf("cannot Parse Json File %s", filePath))
	}
	_ = json.Unmarshal(raw, &w)
	return w
}

func ParseConfig(filePath string) Config {
	var c Config
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		panic(fmt.Errorf("cannot Parse Json File %s", filePath))
	}
	_ = json.Unmarshal(raw, &c)
	return c
}

func LoadConfig(filePath string) Config {
	c := ParseConfig(filePath)
	return c
}

func (wi *WorkflowInstance) RegisterWorkflowDefinition(f func() Workflow) {
	w := f()
	wi.Workflow = w
}

func (wi *WorkflowInstance) RegisterSteps(f func() map[string]func(kubeconfig *string, objName string)) {
	stepFuncMap := f()
	wi.StepFunc = stepFuncMap
}

func (wi *WorkflowInstance) RegisterTrigger(f func(event Event) bool) {
	trigger := f
	wi.Trigger = trigger
}

func (app *App) RegisterWorkflow(definition func() Workflow, steps func() map[string]func(kubeconfig *string, objName string), trigger func(event Event) bool) {
	workflowInstance := CreateWorkflowInstance()
	workflowInstance.RegisterWorkflowDefinition(definition)
	workflowInstance.RegisterSteps(steps)
	workflowInstance.RegisterTrigger(trigger)
	//todo change to relative path
	config := LoadConfig("/Users/gaoxindai/go/src/workflow-engine/workflow-example/config.json")
	app.ModelGVRMap = config.GVRMap
	app.WorkflowInstances = append(app.WorkflowInstances, workflowInstance)
}

func (app *App) Start() {
	fmt.Println("Started Watching")
	kubeconfig := util.GetKubeConfig()
	namespace := "default"
	for _, element := range app.ModelGVRMap {
		ch := util.WatchObject(kubeconfig, namespace, element.Group, element.Version, element.Resource)
		for event := range ch {
			d := event.Object.(*unstructured.Unstructured)
			objKind, found, err := unstructured.NestedString(d.Object, "kind")
			if err != nil || !found {
				fmt.Printf("kind not found for %s %s: error=%s", element.Resource, d.GetName(), err)
				continue
			}
			e := Event{
				Type:  string(event.Type),
				Model: strings.ToLower(objKind),
			}
			for _, wi := range app.WorkflowInstances {
				if wi.Trigger(e) {
					wfObjName := util.GenerateWorkflowObjName()
					util.CreateEmptyWorkflowObject(kubeconfig, wfObjName)
					wi.ExecuteWorkflow(kubeconfig, wfObjName)
				}
			}
		}
	}
}
