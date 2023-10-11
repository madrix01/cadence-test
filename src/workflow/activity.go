package workflow

import (
	"context"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

func SimpleActivity(ctx context.Context, value string) (string, error) {
	activity.GetLogger(ctx).Info("SimpleActivity called.", zap.String("Value", value))
	return "Processed activity: " + value, nil
}
