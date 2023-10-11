package worker

import (
	"cadtest/src/utils"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func buildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to build logger.")
	}

	return logger
}

func buildCadenceServiceClient(config *utils.Config) workflowserviceclient.Interface {
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: config.ClientName,
		Outbounds: yarpc.Outbounds{
			config.CadenceService: {Unary: grpc.NewTransport().NewSingleOutbound(config.HostPort)},
		},
	})

	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	clientConfig := dispatcher.ClientConfig(config.CadenceService)
	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)
}

func StartWorker() worker.Worker {
	config := utils.GetConfig()
	service := buildCadenceServiceClient(config)
	logger := buildLogger()

	workerOptions := worker.Options{
		Logger: logger,
	}
	worker := worker.New(service, config.Domain, config.TaskListName, workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start worker.")
	}

	logger.Info("Started Worker.", zap.String("worker", config.TaskListName))
	return worker
}
