package workflows

import (
	"context"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
)

const (
	TaskListName = "batcherTask"
	batchSize               = 10
	processingTimeThreshold = time.Second * 300
	signalName              = "batcherSignal"
)

func init() {
	workflow.RegisterWithOptions(BatcherWorkflow, workflow.RegisterOptions{
		Name: "BatcherWorkflow",
	})
	activity.RegisterWithOptions(ProcessCustomerActivity, activity.RegisterOptions{
		Name: "ProcessCustomerActivity",
	})
}


func ProcessCustomerActivity(ctx context.Context, batchedCustomerIds []string) ([]string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Running test activity")
	time.Sleep(time.Second * 30)
	return batchedCustomerIds, nil
}

func BatcherWorkflow(ctx workflow.Context, batchedCustomerIds []string) (error) {
	return nil
}

func setupSignalHandler(ctx workflow.Context, selector *workflow.Selector, signalChan *workflow.Channel, signalVal *string) {
}

func setupTimerHandler(ctx workflow.Context, selector *workflow.Selector, timerExpired *bool) {
}
