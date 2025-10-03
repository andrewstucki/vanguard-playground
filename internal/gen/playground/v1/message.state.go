package playgroundv1

import (
	"time"

	"github.com/andrewstucki/protoc-states/workflows"
)

const SendMessageStateWorkflow = "SendMessageState"

type SendMessageStateWorkflowHandler interface {
	Do(io *SendMessageState) error
}

func workflowStepDo(handler SendMessageStateWorkflowHandler) *workflows.WorkflowStep[SendMessageState] {
	return &workflows.WorkflowStep[SendMessageState]{
		Name: "do",
		Fn:   handler.Do,
		Retries: &workflows.RetryPolicy{
			MaxAttempts:          5,
			InitialRetryInterval: 1 * time.Second,
			BackoffCoefficient:   2,
			MaxRetryInterval:     10 * time.Second,
			RetryTimeout:         60 * time.Second,
		},
	}
}

func NewSendMessageStateWorkflowRegistration(handler SendMessageStateWorkflowHandler) workflows.Registration {
	return workflows.NewRegistration(&workflows.Workflow[SendMessageState]{
		Name:       SendMessageStateWorkflow,
		Entrypoint: workflowStepDo(handler),
	})
}
