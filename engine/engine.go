package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"workflow-engine/util"
)

type StepData struct {
	Code    string `json:"code"`
	IsAsync bool   `json:"isAsync"`
}

type StepProps struct {
	Label    string `json:"label"`
	Type     string `json:"type"`
	Group    string `json:"group"`
	Category string `json:"category"`
}

type StepCondition struct {
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Operator   string   `json:"operator"`
	StepLabels []string `json:"stepLabels"`
}

type Step struct {
	Props     StepProps       `json:"props"`
	Data      StepData        `json:"data"`
	NextSteps []StepCondition `json:"nextSteps"`
}

type Workflow struct {
	Name    string          `json:"name"`
	StartAt string          `json:"startAt"`
	Steps   map[string]Step `json:"steps"`
}

type WorkflowInstance struct {
	Workflow Workflow
	StepFunc map[string]func() string
	//todo trigger
	Trigger string
}

func CreateTask() WorkflowInstance {
	var wi WorkflowInstance
	return wi
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

func (wi *WorkflowInstance) RegisterWorkflowDefinition(f func() Workflow) {
	w := f()
	wi.Workflow = w
}

func (wi *WorkflowInstance) RegisterSteps(f func() map[string]func() string) {
	stepFuncMap := f()
	wi.StepFunc = stepFuncMap
}

//todo RegisterTrigger
func (wi *WorkflowInstance) RegisterTrigger(f func() string) {
	trigger := f()
	wi.Trigger = trigger
}

func (wi *WorkflowInstance) Listen() {
	fmt.Println("Started Watching")
	kubeconfig := util.GetKubeConfig()
	ch := util.WatchObject(kubeconfig, "default", "flint.flint.com", "v1", "expenses")
	for event := range ch {
		fmt.Println(event)
		switch event.Type {
		case "ADDED":
			util.CreateSampleWorkflowObject(kubeconfig)
			wi.ExecuteWorkflow(kubeconfig, "default", "flint.flint.com", "v1", "workflows")
		}
		//d := event.Object.(*unstructured.Unstructured)
		//user, found, err := unstructured.NestedString(d.Object, "spec", "user")
		//if err != nil || !found {
		//	fmt.Printf("user not found for expense %s: error=%s", d.GetName(), err)
		//	continue
		//}
		//fmt.Printf("expense %s username is %s\n", d.GetName(), user)
	}
}
