package workflow

import (
	"cadtest/src/utils"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

func SimpleWorkflow(ctx workflow.Context, value string, config *utils.Config) error {
	ao := workflow.ActivityOptions{
		TaskList:               config.ApplicationName,
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 10,
		WaitForCancellation:    false,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	future := workflow.ExecuteActivity(ctx, SimpleActivity, value)
	var result string
	if err := future.Get(ctx, &result); err != nil {
		return err
	}
	workflow.GetLogger(ctx).Info("Done", zap.String("result", result))
	return nil
}
