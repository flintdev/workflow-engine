package engine

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/handler/flowdata"
	"github.com/flintdev/workflow-engine/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	Inputs      []string             `json:"inputs"`
	Condition   string               `json:"condition"`
	NextSteps   []NextStep           `json:"nextSteps"`
}

type Workflow struct {
	Name    string           `json:"name"`
	StartAt string           `json:"startAt"`
	Trigger TriggerCondition `json:"trigger"`
	Steps   map[string]Step  `json:"steps"`
}

type WorkflowInstance struct {
	Workflow     Workflow
	StepTriggers []StepTriggerCondition
}

type App struct {
	WorkflowInstances []WorkflowInstance
	ModelGVRMap       map[string]GVR
	StartAt           time.Time
}

type Event struct {
	Type    string
	Model   string
	Kind    string
	Name    string
	Version string
	Object  interface{}
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

func (wi *WorkflowInstance) RegisterWorkflowDefinition(f func() Workflow) {
	w := f()
	for stepName, step := range w.Steps {
		if step.Type == "manual" {
			stepTrigger := w.Steps[stepName].StepTrigger
			stepTrigger.StepName = stepName
			wi.StepTriggers = append(wi.StepTriggers, stepTrigger)
		}
	}
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

func ParseTrigger(t TriggerCondition, e Event) (bool, error) {
	whenExpresionResult := true
	if t.When != "" {
		result, err := ParseTriggerCondition(t.When, e)
		if err != nil {
			return false, err
		}
		whenExpresionResult = result
	}

	if e.Model == t.Model && strings.ToLower(e.Type) == strings.ToLower(t.EventType) && whenExpresionResult {
		return true, nil
	} else {
		return false, nil
	}
}

func ParseStepTrigger(st StepTriggerCondition, e Event) (bool, error) {
	whenExpresionResult := true
	if st.When != "" {
		result, err := ParseTriggerCondition(st.When, e)
		if err != nil {
			return false, err
		}
		whenExpresionResult = result
	}

	if e.Model == st.Model && strings.ToLower(e.Type) == strings.ToLower(st.EventType) && whenExpresionResult {
		return true, nil
	} else {
		return false, nil
	}
}

func ParseTriggerCondition(input string, e Event) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(input)
	if err != nil {
		return false, err
	}

	var varTokenSlice []interface{}

	tokens := expression.Tokens()
	for i := 0; i < len(tokens); i += 3 {
		varTokenSlice = append(varTokenSlice, tokens[i].Value)
	}

	parameters := make(map[string]interface{})

	for _, token := range varTokenSlice {
		tokenValue := token.(string)
		parsedTokenValue := strings.Replace(tokenValue, ".", "_", -1)
		filedValue, err := getFiledValueByJsonPath(e, tokenValue)
		if err != nil {
			return false, err
		}
		parameters[parsedTokenValue] = filedValue
		input = strings.Replace(input, "'"+tokenValue+"'", parsedTokenValue, -1)
	}
	newExpression, err := govaluate.NewEvaluableExpression(input)
	if err != nil {
		return false, err
	}

	output, err := newExpression.Evaluate(parameters)
	if err != nil {
		return false, err
	}
	if output == nil {
		return false, errors.New("failed to evaluate expression input")
	}
	result := output.(bool)
	return result, nil
}

func getFiledValueByJsonPath(e Event, fieldInput string) (string, error) {
	j := jsonpath.New(uuid.New().String())
	j.AllowMissingKeys(false)
	field, err := get.RelaxedJSONPathExpression(fieldInput)
	if err != nil {
		return "", err
	}
	err = j.Parse(field)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = j.Execute(buf, e.Object)
	if err != nil {
		return "", err
	}
	out := buf.String()
	return out, nil
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
	logger, _ := zap.NewProduction()
	sugarLogger := logger.Sugar()
	defer sugarLogger.Sync()
	for event := range ch {
		sugarLogger.Infow("Received Event",
			"Type", event.Type,
			"Object", event.Object,
		)
		d := event.Object.(*unstructured.Unstructured)
		objKind := d.GetKind()
		objName := d.GetName()
		objVersion := d.GetAPIVersion()
		creationTimestamp, _, _ := unstructured.NestedString(d.Object, "metadata", "creationTimestamp")
		t, err := time.Parse(time.RFC3339, creationTimestamp)
		if err != nil {
			message := fmt.Sprintf("cannot parse timestamp %s: error=%s", creationTimestamp, err)
			sugarLogger.Error(message,
				"Type", event.Type,
				"Object", event.Object,
			)
			continue
		}
		if app.StartAt.After(t) && event.Type == "ADDED" {
			continue
		}
		e := Event{
			Type:    string(event.Type),
			Model:   strings.ToLower(objKind),
			Object:  event.Object.(*unstructured.Unstructured).Object,
			Kind:    objKind,
			Name:    objName,
			Version: objVersion,
		}
		for _, wi := range app.WorkflowInstances {
			go triggerWorkflowInstance(kubeconfig, objName, wi, e)
		}
	}
}

func triggerWorkflowInstance(kubeconfig *string, objName string, wi WorkflowInstance, e Event) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	isWorkflowTriggered, err := util.CheckIfWorkflowIsTriggered(kubeconfig, objName)
	if err != nil {
		logger.Error(err.Error(),
			zap.String("Kind", e.Kind),
			zap.String("Name", e.Name),
			zap.String("Version", e.Version),
		)
	}
	if isWorkflowTriggered {
		for _, stepTrigger := range wi.StepTriggers {
			result, err := ParseStepTrigger(stepTrigger, e)
			if err != nil {
				logger.Warn(err.Error(),
					zap.String("Kind", e.Kind),
					zap.String("Name", e.Name),
					zap.String("Version", e.Version),
					zap.String("Trigger condition", stepTrigger.When),
				)
				continue
			}
			if result {
				currentStep := stepTrigger.StepName
				objList, err := util.GetPendingWorkflowList(kubeconfig, objName, currentStep)
				if err != nil {
					logger.Error(err.Error(),
						zap.String("Kind", e.Kind),
						zap.String("Name", e.Name),
						zap.String("Version", e.Version),
					)
					continue
				}
				for _, obj := range objList.Items {
					wfObjName := obj.GetName()
					var fd flowdata.FlowData
					fd.Kubeconfig = kubeconfig
					fd.WFObjName = wfObjName
					var h handler.Handler
					h.FlowData = fd
					wi.ExecuteWorkflow(kubeconfig, logger, h, wfObjName, currentStep, true)
				}
			}
		}
	} else {
		result, err := ParseTrigger(wi.Workflow.Trigger, e)
		if err != nil {
			logger.Warn(err.Error(),
				zap.String("Kind", e.Kind),
				zap.String("Name", e.Name),
				zap.String("Version", e.Version),
			)
		}
		if result {
			startAt := wi.Workflow.StartAt
			wfObjName := util.GenerateWorkflowObjName()
			var fd flowdata.FlowData
			fd.Kubeconfig = kubeconfig
			fd.WFObjName = wfObjName
			var h handler.Handler
			h.FlowData = fd
			err := util.CreateEmptyWorkflowObject(kubeconfig, wfObjName, objName)
			if err != nil {
				logger.Error(err.Error())
			} else {
				wi.ExecuteWorkflow(kubeconfig, logger, h, wfObjName, startAt, false)
			}

		}
	}
}

func BulkWatchObject(kubeconfig *string, namespace string, gvrList []GVR) <-chan watch.Event {
	var chans []<-chan watch.Event
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	for _, gvr := range gvrList {
		message := fmt.Sprintf("Start Watching Resource Group: %s, Version: %s, Resource: %s", gvr.Group, gvr.Version, gvr.Resource)
		logger.Info(message)
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
