package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
	"io/ioutil"
	"net/http"
)

type ExecutorResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func executeStep(kubeconfig *string, wi *WorkflowInstance, handler handler.Handler, wfObjName string, currentStep string) (bool, string, error) {
	wfMessage := fmt.Sprintf("running step %s", currentStep)
	err := util.SetWorkflowObjectMessage(kubeconfig, wfObjName, wfMessage)
	if err != nil {
		return false, "", err
	}
	emptyCondition := StepCondition{}
	nextStep := ""
	port := getPythonExecutorPort()
	url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, currentStep, wfObjName)
	fmt.Printf("Sent GET request to %s\n", url)
	response, err := http.Get(url)
	if err != nil {
		message := fmt.Sprintf("The HTTP request failed with error %s\n", err)
		return false, "", errors.New(message)
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false, "", err
		}
		r, err := ParseExecutorResponse(data)
		if err != nil {
			return false, "", err
		}
		if r.Status == "success" {
			nextSteps := wi.Workflow.Steps[currentStep].NextSteps
			for _, step := range nextSteps {
				nextStepName := step.Name
				condition := step.Condition
				if condition == emptyCondition {
					nextStep = nextStepName
					break
				}
				key := condition.Key
				value := condition.Value
				operator := condition.Operator
				flowDataResult, err := handler.FlowData.Get(key)
				if err != nil {
					return false, "", err
				}
				switch operator {
				case "=":
					if value == flowDataResult {
						nextStep = nextStepName
						break
					}
				}
			}
			err := util.SetWorkflowObjectStepToComplete(kubeconfig, wfObjName, currentStep)
			if err != nil {
				return false, "", err
			}
			if len(wi.Workflow.Steps[nextStep].NextSteps) == 0 {
				return true, "", nil
			} else {
				return false, nextStep, nil
			}
		} else {
			err := util.SetWorkflowObjectFailedStep(kubeconfig, wfObjName, currentStep, r.Message)
			if err != nil {
				return false, "", err
			}
			message := fmt.Sprintf("Failed on step %s. Reason: %s", currentStep, r.Message)
			return false, "", errors.New(message)
		}
	}
}

func completeWorkflow (kubeconfig *string, wfObjName string, wi *WorkflowInstance) error {
	err := util.SetWorkflowObjectToComplete(kubeconfig, wfObjName)
	if err !=nil {
		return err
	}
	wfMessage := fmt.Sprintf("Workflow %s Complete\n", wi.Workflow.Name)
	fmt.Printf(wfMessage)
	err = util.SetWorkflowObjectMessage(kubeconfig, wfObjName, wfMessage)
	if err != nil {
		return err
	}
	err = util.SetWorkflowObjectCurrentStep(kubeconfig, wfObjName, "end")
	if err != nil {
		return err
	}
	err = util.SetWorkflowObjectCurrentStepLabel(kubeconfig, wfObjName, "end")
	return nil
}

func handleManualStep (kubeconfig *string, wfObjName string, wi *WorkflowInstance, currentStep string) error{
	err := util.SetWorkflowObjectStepToPending(kubeconfig, wfObjName, currentStep)
	if err != nil {
		return err
	}
	stepTrigger := wi.Workflow.Steps[currentStep].StepTrigger
	stepTrigger.StepName = currentStep
	wi.StepTriggers = append(wi.StepTriggers, stepTrigger)
	wfMessage := fmt.Sprintf("Workflow is pending on manual step %s\n", currentStep)
	fmt.Printf(wfMessage)
	err = util.SetWorkflowObjectMessage(kubeconfig, wfObjName, wfMessage)
	if err != nil {
		return err
	}
	return nil
}

func (wi *WorkflowInstance) ExecuteWorkflow(kubeconfig *string, handler handler.Handler, wfObjName string) error {
	currentStep := ""
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	err := util.SetWorkflowObjectToRunning(kubeconfig, wfObjName)
	if err !=nil {
		err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
		if err != nil {
			return err
		}
	}
	for {
		fmt.Printf("Start Running Step %s\n", currentStep)
		err := util.SetWorkflowObjectStep(kubeconfig, wfObjName, currentStep)
		if err != nil {
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				return err
			}
		}
		if wi.Workflow.Steps[currentStep].Type == "manual" {
			err := handleManualStep(kubeconfig, wfObjName, wi, currentStep)
			if err != nil {
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					return err
				}
			}
			break
		}
		isComplete, nextStep, err := executeStep(kubeconfig, wi, handler, wfObjName, currentStep)
		currentStep = nextStep
		if err != nil {
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				return err
			}
			return err
		}
		if isComplete {
			err := completeWorkflow(kubeconfig, wfObjName, wi)
			if err != nil {
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					return err
				}
				break
			}
			break
		}
	}
	return nil
}

func (wi *WorkflowInstance) ExecutePendingWorkflow(kubeconfig *string, handler handler.Handler, wfObjName string, currentStep string) error{
	fmt.Printf("Start Executing Pending Workflow %s\n", wi.Workflow.Name)
	skipFirst := false
	for {
		fmt.Printf("Start Running Step %s\n", currentStep)
		err := util.SetWorkflowObjectStepToRunning(kubeconfig, wfObjName, currentStep)
		if err != nil {
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				return err
			}
		}
		if skipFirst {
			err := util.SetWorkflowObjectStep(kubeconfig, wfObjName, currentStep)
			if err != nil {
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					return err
				}
			}
			if wi.Workflow.Steps[currentStep].Type == "manual" {
				err := handleManualStep(kubeconfig, wfObjName, wi, currentStep)
				if err != nil {
					err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
					if err != nil {
						return err
					}
				}
				break
			}
		}
		skipFirst = true
		isComplete, nextStep, err := executeStep(kubeconfig, wi, handler, wfObjName, currentStep)
		if err != nil {
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				return err
			}
			return err
		}
		currentStep = nextStep
		if isComplete {
			err := completeWorkflow(kubeconfig, wfObjName, wi)
			if err != nil {
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					return err
				}
				break
			}
			break
		}
	}
	return nil
}

func ParseExecutorResponse(body []byte) (ExecutorResponse, error) {
	var r ExecutorResponse
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func getPythonExecutorPort() string {
	data, _ := ioutil.ReadFile("/tmp/flint_python_executor_port")
	port := string(data)
	return port
}
