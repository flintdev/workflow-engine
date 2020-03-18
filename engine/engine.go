package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/handler/flowdata"
	"github.com/flintdev/workflow-engine/util"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/kubectl/pkg/cmd/get"
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

type TriggerCondition struct {
	Model     string `json:"model"`
	EventType string `json:"eventType"`
	When      string `json:"when"`
}

type StepTriggerCondition struct {
	StepName  string `json:"stepName"`
	Model     string `json:"model"`
	EventType string `json:"eventType"`
	When      string `json:"when"`
}

type Step struct {
	Type        string               `json:"type"`
	StepTrigger StepTriggerCondition `json:"trigger"`
	NextSteps   []NextStep           `json:"nextSteps"`
}

type Workflow struct {
	Name    string           `json:"name"`
	StartAt string           `json:"startAt"`
	Trigger TriggerCondition `json:"trigger"`
	Steps   map[string]Step  `json:"steps"`
}

// todo remove all non-instance related field
type WorkflowInstance struct {
	Workflow     Workflow
	Kubeconfig   *string
	ModelObjName string
	StepTriggers []StepTriggerCondition
}

type App struct {
	WorkflowInstances []WorkflowInstance
	ModelGVRMap       map[string]GVR
	StartAt           time.Time
}

type Event struct {
	Type   string
	Model  string
	Object interface{}
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
	fmt.Println(t.When)
	whenExpresionResult := true
	if t.When != "" {
		whenExpresionResult = ParseTriggerCondition(t, e)
	}

	if e.Model == t.Model && strings.ToLower(e.Type) == strings.ToLower(t.EventType) && whenExpresionResult {
		return true
	} else {
		return false
	}
}

func ParseStepTrigger(st StepTriggerCondition, e Event) bool {
	fmt.Println(st.When)
	whenExpresionResult := true
	if st.When != "" {
		whenExpresionResult = ParseStepTriggerCondition(st, e)
	}

	if e.Model == st.Model && strings.ToLower(e.Type) == strings.ToLower(st.EventType) && whenExpresionResult {
		return true
	} else {
		return false
	}
}

func ParseTriggerCondition(t TriggerCondition, e Event) bool {
	input := t.When
	expression, _ := govaluate.NewEvaluableExpression(input)
	var varTokenSlice []interface{}

	tokens := expression.Tokens()
	for i := 0; i < len(tokens); i += 3 {
		varTokenSlice = append(varTokenSlice, tokens[i].Value)
	}

	parameters := make(map[string]interface{})

	for _, token := range varTokenSlice {
		tokenValue := token.(string)
		fmt.Println(tokenValue)
		parsedTokenValue := strings.Replace(tokenValue, ".", "_", -1)
		fmt.Println(parsedTokenValue)
		filedValue := getFiledValueByJsonPath(e, tokenValue)
		fmt.Println(filedValue)
		parameters[parsedTokenValue] = filedValue
		input = strings.Replace(input, "'"+tokenValue+"'", parsedTokenValue, -1)
	}
	fmt.Println(input)
	fmt.Println(parameters)
	newExpression, _ := govaluate.NewEvaluableExpression(input)

	output, err := newExpression.Evaluate(parameters)
	fmt.Println(err)
	result := output.(bool)
	fmt.Println(result)
	return result
}

func ParseStepTriggerCondition(st StepTriggerCondition, e Event) bool {
	input := st.When
	expression, _ := govaluate.NewEvaluableExpression(input)
	var varTokenSlice []interface{}

	tokens := expression.Tokens()
	for i := 0; i < len(tokens); i += 3 {
		varTokenSlice = append(varTokenSlice, tokens[i].Value)
	}

	parameters := make(map[string]interface{})

	for _, token := range varTokenSlice {
		tokenValue := token.(string)
		fmt.Println(tokenValue)
		parsedTokenValue := strings.Replace(tokenValue, ".", "_", -1)
		fmt.Println(parsedTokenValue)
		filedValue := getFiledValueByJsonPath(e, tokenValue)
		fmt.Println(filedValue)
		parameters[parsedTokenValue] = filedValue
		input = strings.Replace(input, "'"+tokenValue+"'", parsedTokenValue, -1)
	}
	fmt.Println(input)
	fmt.Println(parameters)
	newExpression, _ := govaluate.NewEvaluableExpression(input)

	output, err := newExpression.Evaluate(parameters)
	fmt.Println(err)
	result := output.(bool)
	fmt.Println(result)
	return result
}

func getFiledValueByJsonPath(e Event, fieldInput string) string {
	j := jsonpath.New(uuid.New().String())
	j.AllowMissingKeys(false)
	field, err := get.RelaxedJSONPathExpression(fieldInput)
	ee := j.Parse(field)
	fmt.Println(ee)
	buf := new(bytes.Buffer)
	err = j.Execute(buf, e.Object)
	fmt.Println(err)
	out := buf.String()
	fmt.Println(out)
	return out
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
		objName, found, objNameERR := unstructured.NestedString(d.Object, "metadata", "name")
		if objKindERR != nil || !found {
			fmt.Printf("kind not found for %s %s: error=%s\n", objKind, d.GetName(), objKindERR)
			continue
		}
		if creationTimestampERR != nil || !found {
			fmt.Printf("creationTimestamp not found for %s %s: error=%s\n", objKind, d.GetName(), creationTimestampERR)
			continue
		}
		if objNameERR != nil || !found {
			fmt.Printf("objNameERR not found for %s %s: error=%s\n", objKind, d.GetName(), creationTimestampERR)
			continue
		}
		t, err := time.Parse(time.RFC3339, creationTimestamp)
		if err != nil {
			fmt.Printf("cannot parse timestamp %s: error=%s\n", creationTimestamp, err)
			continue
		}
		if app.StartAt.After(t) && event.Type == "ADDED" {
			continue
		}
		e := Event{
			Type:   string(event.Type),
			Model:  strings.ToLower(objKind),
			Object: event.Object.(*unstructured.Unstructured).Object,
		}
		for index, wi := range app.WorkflowInstances {
			fmt.Println(index)
			fmt.Println(wi.StepTriggers)
			fmt.Println("filter workflow")
			if util.CheckIfWorkflowIsTriggered(kubeconfig, objName) {
				fmt.Println("Printing Step trigger")
				fmt.Println(wi.StepTriggers)
				for _, stepTrigger := range wi.StepTriggers {
					fmt.Println(stepTrigger)
					if ParseStepTrigger(stepTrigger, e) {
						fmt.Println("ready to start pending step")
						currentStep := stepTrigger.StepName
						objList := util.GetPendingWorkflowList(kubeconfig, objName, currentStep)
						for _, obj := range objList.Items {
							wi.Kubeconfig = kubeconfig
							wfObjName := obj.GetName()
							fmt.Println(wfObjName)
							fmt.Println(currentStep)
							var fd flowdata.FlowData
							fd.Kubeconfig = kubeconfig
							fd.WFObjName = wfObjName
							var h handler.Handler
							h.FlowData = fd
							wi.ExecutePendingWorkflow(h, wfObjName, currentStep)
						}
					}
				}
			} else {
				if ParseTrigger(wi.Workflow.Trigger, e) {
					fmt.Println(kubeconfig)
					wi.ModelObjName = objName
					wfObjName := util.GenerateWorkflowObjName()
					wi.Kubeconfig = kubeconfig
					var fd flowdata.FlowData
					fd.Kubeconfig = kubeconfig
					fd.WFObjName = wfObjName
					var h handler.Handler
					h.FlowData = fd
					util.CreateEmptyWorkflowObject(kubeconfig, wfObjName, objName)
					wi.ExecuteWorkflow(h, wfObjName)
					fmt.Println("-------------")
					fmt.Println(wi.StepTriggers)
					app.WorkflowInstances[index].StepTriggers = wi.StepTriggers
					break
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
