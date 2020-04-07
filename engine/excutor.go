package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
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
			if len(nextSteps) == 0 {
				err = checkAllExistingStepsStatus(kubeconfig, logger, wfObjName, stepName)
				if err != nil {
					return
				}
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
			err = checkAllExistingStepsStatus(kubeconfig, logger, wfObjName, stepName)
			if err != nil {
				return
			}
		} else {
			for _, step := range nextSteps {
				go executeStep(kubeconfig, wi, logger, port, wfObjName, step.Name, handler, false)
			}
		}
	}
}

// check all existing steps status.
func checkAllExistingStepsStatus(kubeconfig *string, logger *zap.Logger, wfObjName string, stepName string) error {
	status, message, err := util.CheckAllStepStatus(kubeconfig, wfObjName)
	if err != nil {
		logError(logger, wfObjName, stepName, err.Error())
		err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			return err
		}
		return err
	}
	switch status {
	case "allCompleteSuccess":
		err := util.SetWorkflowObjectToComplete(kubeconfig, wfObjName)
		if err != nil {
			logError(logger, wfObjName, stepName, err.Error())
			err := util.SetWorkflowObjectToFailure(kubeconfig, wfObjName, err.Error())
			if err != nil {
				logError(logger, wfObjName, stepName, err.Error())
				return err
			}
			return err
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
				return err
			}
			return err
		}
	}
	return nil
}

// get next steps by a given step.
func getNextSteps(wi *WorkflowInstance, r ExecutorResponse, stepName string, handler handler.Handler) ([]NextStep, error) {
	var nextMatchedSteps []NextStep
	if r.Status == "success" {
		nextSteps := wi.Workflow.Steps[stepName].NextSteps
		for _, step := range nextSteps {
			if step.When == "" {
				nextMatchedSteps = append(nextMatchedSteps, step)
				continue
			}
			result, err := parseStepCondition(step.When, handler)
			if err != nil {
				return nextMatchedSteps, err
			}
			if result {
				nextMatchedSteps = append(nextMatchedSteps, step)
			}
		}
		return nextMatchedSteps, nil
	} else {
		message := fmt.Sprintf(r.Message)
		return nextMatchedSteps, errors.New(message)
	}
}

//parse step condition
func parseStepCondition(input string, handler handler.Handler) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(input)
	if err != nil {
		return false, err
	}

	var varTokenSlice []interface{}

	tokens := expression.Tokens()
	for i := 0; i < len(tokens); i += 4 {
		varTokenSlice = append(varTokenSlice, tokens[i].Value)
	}
	parameters := make(map[string]interface{})

	for _, token := range varTokenSlice {
		tokenValue := token.(string)
		parsedTokenValue := util.ParseFlowDataKey(tokenValue)
		parsedTokenValue = strings.Replace(parsedTokenValue, ".", "_", -1)
		flowDataResult, err := handler.FlowData.Get(tokenValue)
		if err != nil {
			return false, err
		}
		switch flowDataResult.(type) {
		case string:
			input = strings.Replace(input, "'"+tokenValue+"'", "'"+flowDataResult.(string)+"'", -1)
		default:
			parameters[parsedTokenValue] = flowDataResult
			input = strings.Replace(input, "'"+tokenValue+"'", parsedTokenValue, -1)
		}
	}
	newExpression, err := govaluate.NewEvaluableExpression(input)
	if err != nil {
		return false, err
	}

	output, err := newExpression.Evaluate(parameters)
	fmt.Println(input)
	fmt.Println(parameters)
	fmt.Println(output)
	if err != nil {
		return false, err
	}
	switch output.(type) {
	case bool:
		result := output.(bool)
		return result, nil
	default:
		return false, errors.New("failed to evaluate expression input")
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
