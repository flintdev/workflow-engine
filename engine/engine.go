package engine

import (
	"encoding/json"
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/handler/flowdata"
	"github.com/flintdev/workflow-engine/util"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"sync"
	"time"
)

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

type TriggerCondition struct {
	Model     string `json:"model"`
	EventType string `json:"eventType"`
	When      string `json:"when"`
}

type Workflow struct {
	Name    string               `json:"name"`
	StartAt string               `json:"startAt"`
	Trigger TriggerCondition     `json:"trigger"`
	Steps   map[string]NextSteps `json:"steps"`
}

type WorkflowInstance struct {
	Workflow   Workflow
	Kubeconfig *string
	WFObjName  string
}

type App struct {
	WorkflowInstances []WorkflowInstance
	ModelGVRMap       map[string]GVR
	StartAt           time.Time
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

func (wi *WorkflowInstance) RegisterWorkflowDefinition(f func() Workflow) {
	w := f()
	wi.Workflow = w
}

func (app *App) RegisterConfig(f func() Config) {
	c := f()
	app.ModelGVRMap = c.GVRMap
}

func (app *App) RegisterWorkflow(definition func() Workflow) {
	workflowInstance := CreateWorkflowInstance()
	workflowInstance.RegisterWorkflowDefinition(definition)
	app.WorkflowInstances = append(app.WorkflowInstances, workflowInstance)
}

func ParseTrigger(t TriggerCondition, e Event) bool {
	if e.Model == t.Model && strings.ToLower(e.Type) == strings.ToLower(t.EventType) {
		return true
	} else {
		return false
	}
}

func (app *App) Start() {
	kubeconfig := util.GetKubeConfig()
	namespace := "default"
	app.StartAt = time.Now()
	var gvrList []GVR
	for _, element := range app.ModelGVRMap {
		gvrList = append(gvrList, element)
	}
	ch := BulkWatchObject(kubeconfig, namespace, gvrList)
	triggerWorkflow(kubeconfig, ch, app)
}

func triggerWorkflow(kubeconfig *string, ch <-chan watch.Event, app *App) {
	for event := range ch {
		fmt.Println("Received Event")
		fmt.Println("Event Type:", event.Type)
		fmt.Println("Event Object", event.Object)
		d := event.Object.(*unstructured.Unstructured)
		creationTimestamp, found, objKindERR := unstructured.NestedString(d.Object, "metadata", "creationTimestamp")
		objKind, found, creationTimestampERR := unstructured.NestedString(d.Object, "kind")
		if objKindERR != nil || !found {
			fmt.Printf("kind not found for %s %s: error=%s\n", objKind, d.GetName(), objKindERR)
			continue
		}
		if creationTimestampERR != nil || !found {
			fmt.Printf("creationTimestamp not found for %s %s: error=%s\n", objKind, d.GetName(), creationTimestampERR)
			continue
		}
		t, err := time.Parse(time.RFC3339, creationTimestamp)
		if err != nil {
			fmt.Printf("cannot parse timestamp %s: error=%s\n", creationTimestamp, err)
			continue
		}
		if app.StartAt.Before(t) {
			e := Event{
				Type:  string(event.Type),
				Model: strings.ToLower(objKind),
			}
			for _, wi := range app.WorkflowInstances {
				if ParseTrigger(wi.Workflow.Trigger, e) {
					wfObjName := util.GenerateWorkflowObjName()
					wi.WFObjName = wfObjName
					wi.Kubeconfig = kubeconfig
					var fd flowdata.FlowData
					fd.Kubeconfig = kubeconfig
					fd.WFObjName = wfObjName
					var h handler.Handler
					h.FlowData = fd
					util.CreateEmptyWorkflowObject(kubeconfig, wfObjName)
					wi.ExecuteWorkflow(h)
				}
			}
		}
	}
}

func BulkWatchObject(kubeconfig *string, namespace string, gvrList []GVR) <-chan watch.Event {
	var chans []<-chan watch.Event
	for _, gvr := range gvrList {
		fmt.Printf("Start Watching Resource Group: %s, Version: %s, Resource: %s\n", gvr.Group, gvr.Version, gvr.Resource)
		chans = append(chans, util.WatchObject(kubeconfig, namespace, gvr.Group, gvr.Version, gvr.Resource))
	}
	ch := mergeWatchChannels(chans)
	return ch
}

func mergeWatchChannels(chans []<-chan watch.Event) <-chan watch.Event {
	ch := make(chan watch.Event)
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(chans))

		for _, c := range chans {
			go func(c <-chan watch.Event) {
				for v := range c {
					ch <- v
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(ch)
	}()
	return ch
}
