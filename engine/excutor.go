package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type ExecutorResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (wi *WorkflowInstance) ExecuteWorkflow(kubeconfig *string, logger *zap.Logger, handler handler.Handler, wfObjName string, stepName string, isPendingManualStep bool) {
	logInfo(logger, wfObjName, stepName, "Start Executing Workflow")
	port, err := getPythonExecutorPort()
	if err != nil {
		logError(logger, wfObjName, stepName, err.Error())
		err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			return
		}
		return
	}
	err = util.SetWorkflowObjectToRunning(kubeconfig, wfObjName)
	if err != nil {
		logError(logger, wfObjName, stepName, err.Error())
		err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			return
		}
		return
	}
	executeStep(kubeconfig, wi, logger, port, wfObjName, stepName, handler, isPendingManualStep)
}
func executeStep(kubeconfig *string, wi *WorkflowInstance, logger *zap.Logger, port string, wfObjName string, stepName string, handler handler.Handler, isPendingManualStep bool) {
	// handle hub step
	if wi.Workflow.Steps[stepName].Type == "hub" {
		inputStepsList := wi.Workflow.Steps[stepName].Inputs
		condition := wi.Workflow.Steps[stepName].Condition
		hubStatus, err := util.CheckAllStepStatusByList(kubeconfig, wfObjName, inputStepsList)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectFailedStep(kubeconfig, wfObjName, stepName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
		if hubStatus != condition {
			return
		}
	}
	if isPendingManualStep {
		logInfo(logger, wfObjName, stepName, "start running step")
		err := util.SetWorkflowObjectStepToRunning(kubeconfig, wfObjName, stepName, "")
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
		err = util.RemoveWorkflowObjectPendingStepLabel(kubeconfig, wfObjName, stepName)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
	} else {
		// handle manual step. Return and wait for trigger
		if wi.Workflow.Steps[stepName].Type == "manual" {
			logInfo(logger, wfObjName, stepName, "pending on manual step")
			err := util.SetPendingStepToWorkflowObject(kubeconfig, stepName, wfObjName)
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					return
				}
				return
			}
			err = util.SetWorkflowObjectPendingStepLabel(kubeconfig, wfObjName, stepName)
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				err := util.SetWorkflowObjectFailedStep(kubeconfig, wfObjName, stepName, err.Error())
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					return
				}
				return
			}
			return
		}
		logInfo(logger, wfObjName, stepName, "start running step")
		err := util.SetStepToWorkflowObject(kubeconfig, stepName, wfObjName)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
	}

	// start executing step
	url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, stepName, wfObjName)
	message := fmt.Sprintf("Sent GET request to %s", url)
	logInfo(logger, wfObjName, stepName, message)
	response, err := http.Get(url)
	if err != nil {
		message := fmt.Sprintf("The HTTP request failed with error %s", err)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectFailedStep(kubeconfig, wfObjName, stepName, message)
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectFailedStep(kubeconfig, wfObjName, stepName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
		r, err := ParseExecutorResponse(data)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectStepToFailure(kubeconfig, wfObjName, stepName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
		nextSteps, err := getNextSteps(wi, r, stepName, handler)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectStepToFailure(kubeconfig, wfObjName, stepName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
			return
		}
		err = util.SetWorkflowObjectStepToComplete(kubeconfig, wfObjName, stepName, "")
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectStepToFailure(kubeconfig, wfObjName, stepName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return
			}
		}

		// check if next step is end
		if len(nextSteps) == 1 && len(wi.Workflow.Steps[nextSteps[0].Name].NextSteps) == 0 {
			wfStatus, err := util.GetWorkflowObjectStatus(kubeconfig, wfObjName)
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					return
				}
				return
			}
			if wfStatus == "Failure" {
				return
			}
			// check all existing steps status.
			status, message, err := util.CheckAllStepStatus(kubeconfig, wfObjName)
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					return
				}
				return
			}
			switch status {
			case "allCompleteSuccess":
				err := util.SetWorkflowObjectToComplete(kubeconfig, wfObjName)
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
					if err != nil {
						logError(logger, wfObjName, stepName, err.Error())
						return
					}
					return
				}
			case "hasRunning":
			case "hasPending":
			case "allCompleteHasFailure":
				err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, message)
				if err != nil {
					logError(logger, wfObjName, stepName, err.Error())
					err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
					if err != nil {
						logError(logger, wfObjName, stepName, err.Error())
						return
					}
					return
				}
			}
		} else {
			for _, step := range nextSteps {
				go executeStep(kubeconfig, wi, logger, port, wfObjName, step.Name, handler, false)
			}
		}
	}
}

// get next steps by a given step.
func getNextSteps(wi *WorkflowInstance, r ExecutorResponse, stepName string, handler handler.Handler) ([]NextStep, error) {
	var nextMatchedSteps []NextStep
	emptyCondition := StepCondition{}
	if r.Status == "success" {
		nextSteps := wi.Workflow.Steps[stepName].NextSteps
		for _, step := range nextSteps {
			condition := step.Condition
			if condition == emptyCondition {
				nextMatchedSteps = append(nextMatchedSteps, step)
				continue
			}
			key := condition.Key
			value := condition.Value
			operator := condition.Operator
			flowDataResult, err := handler.FlowData.Get(key)
			if err != nil {
				return nextMatchedSteps, err
			}
			switch operator {
			case "=":
				if value == flowDataResult {
					nextMatchedSteps = append(nextMatchedSteps, step)
					continue
				}
			}
		}
		return nextMatchedSteps, nil
	} else {
		message := fmt.Sprintf(r.Message)
		return nextMatchedSteps, errors.New(message)
	}
}

func ParseExecutorResponse(body []byte) (ExecutorResponse, error) {
	var r ExecutorResponse
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func getPythonExecutorPort() (string, error) {
	data, err := ioutil.ReadFile("/tmp/flint_python_executor_port")
	if err != nil {
		return "", err
	}
	port := string(data)
	return port, nil
}

func logInfo(logger *zap.Logger, wfObjName string, stepName string, infoMessage string) {
	logger.Info(infoMessage,
		zap.String("workflow object name", wfObjName),
		zap.String("step name", stepName),
	)
}

func logError(logger *zap.Logger, wfObjName string, stepName string, errMessage string) {
	logger.Error(errMessage,
		zap.String("workflow object name", wfObjName),
		zap.String("step name", stepName),
	)
}
