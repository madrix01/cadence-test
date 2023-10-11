package main

import (
	"cadtest/src/common"
	"cadtest/src/utils"
	"cadtest/src/workflow"
	"fmt"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
)

func main() {
	config := utils.GetConfig()
	h := common.Helper{}

	h.Setup(&common.Configuration{
		DomainName:      config.Domain,
		ServiceName:     config.CadenceFrontendService,
		HostNameAndPort: config.HostPort,
	}, config)

	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       &h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	h.RegisterWorkflowWithAlias(workflow.SimpleWorkflow, "hello_world")
	h.RegisterActivity(workflow.SimpleActivity)

	fmt.Println("#####", h.Config.DomainName)
	h.StartWorkers(h.Config.DomainName, config.ApplicationName, workerOptions)
	// workflowOptions := client.StartWorkflowOptions{
	// 	ID:                              "test_task_" + uuid.New().String(),
	// 	TaskList:                        config.ApplicationName,
	// 	ExecutionStartToCloseTimeout:    5 * time.Minute,
	// 	DecisionTaskStartToCloseTimeout: 5 * time.Minute,
	// }

	// h.StartWorkflow(workflowOptions, "hello_world", "hackerman")
}
