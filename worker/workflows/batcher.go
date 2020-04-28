package workflows

import (
	"context"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
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
	isInitialRun := true
	logger := workflow.GetLogger(ctx)
	logger.Info("Batcher workflow started.",
		zap.Time("StartTime", workflow.Now(ctx)),
		zap.Duration("ProcessingTimeThreshold", processingTimeThreshold))



	err := workflow.SetQueryHandler(ctx, "current_data", func() ([]string, error) {
		return batchedCustomerIds, nil
	})
	if err != nil {
		return err
	}

	// Get Ready to wait for Signal
	var signalVal string
	var timerExpired bool

	selector := workflow.NewSelector(ctx)
	signalChan := workflow.GetSignalChannel(ctx, signalName)
	setupSignalHandler(ctx, &selector, &signalChan, &signalVal)
	setupTimerHandler(ctx, &selector, &timerExpired)

	for endWorkflow := false; !endWorkflow; {
		// If it not initial run, we will stop here and wait for signal
		if !isInitialRun {
			logger.Debug("waiting for signal", zap.String("signalName", signalName))
			selector.Select(ctx) // Wait for signal

			if len(signalVal) > 0 {
				batchedCustomerIds = append(batchedCustomerIds, signalVal)
				signalVal = "" // Reset signal value
			}
		} else {
			logger.Debug("Initial run. Skipping select.")
			isInitialRun = false
		}

		switch {
		case timerExpired && len(batchedCustomerIds) == 0:
			// Reset timer if no customerIds to process
			logger.Info("Resetting timer as there are no customers to process!") // Message to check timer is working correctly
			timerExpired = false
			setupTimerHandler(ctx, &selector, &timerExpired)
		case timerExpired && len(batchedCustomerIds) > 0, len(batchedCustomerIds) > batchSize:
			// If timer expired or if we have enough customers, we should start batching
			logger.Info("Batch limit reached! Processing batch now!")

			// Take batch and send to aml service. We will do a slice here in case from the previous
			// workflow, there already exists more than the batch size of items. Items exceeding batch
			// will be processed in the next run.
			var result []string
			var customersToProcess []string
			if len(batchedCustomerIds) > batchSize {
				customersToProcess = batchedCustomerIds[0:batchSize]
				batchedCustomerIds = batchedCustomerIds[batchSize:]
			} else {
				customersToProcess = batchedCustomerIds
				batchedCustomerIds = []string{}
			}
			ao := workflow.ActivityOptions{
				TaskList:               "test",
				ScheduleToCloseTimeout: time.Minute * 50,
				ScheduleToStartTimeout: time.Minute * 50,
				StartToCloseTimeout:    time.Minute * 50,
				HeartbeatTimeout:       time.Minute * 50,
				WaitForCancellation:    false,
			}
			ctx = workflow.WithActivityOptions(ctx, ao)
			actFuture := workflow.ExecuteActivity(ctx, ProcessCustomerActivity, customersToProcess)
			if err := actFuture.Get(ctx, &result); err != nil {
				return err
			}
			endWorkflow = true
		}
	}

	// Start a new batcher workflow, make sure also to pass over any unhandled signals
	ok := true
	// Ok will be true when there are data available. More will be true when there are more data available.
	// As long as any of them return false.
	for ok {
		ok = signalChan.ReceiveAsync(&signalVal)
		if len(signalVal) > 0 {
			batchedCustomerIds = append(batchedCustomerIds, signalVal)
			signalVal = "" // Reset signal value
		}
	}

	logger.Info("Batch workflow completed! Continuing as new..")
	return workflow.NewContinueAsNewError(ctx, BatcherWorkflow, batchedCustomerIds)
}

func setupSignalHandler(ctx workflow.Context, selector *workflow.Selector, signalChan *workflow.Channel, signalVal *string) {
	(*selector).AddReceive(*signalChan, func(c workflow.Channel, more bool) {
		err := ctx.Err()
		if err != nil {
			workflow.GetLogger(ctx).Error("error while receiving signal", zap.String("signal", signalName), zap.Any("value", signalVal), zap.Error(err))
		}
		c.Receive(ctx, signalVal)
		workflow.GetLogger(ctx).Info("received signal", zap.String("signal", signalName), zap.Any("value", signalVal))
	})
}

func setupTimerHandler(ctx workflow.Context, selector *workflow.Selector, timerExpired *bool) {
	// use timer future to send notification email if processing takes too long
	timerFuture := workflow.NewTimer(ctx, processingTimeThreshold)
	(*selector).AddFuture(timerFuture, func(f workflow.Future) {
		*timerExpired = true
	})
}
