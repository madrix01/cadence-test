package common

import (
	"cadtest/src/utils"
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type (
	Configuration struct {
		DomainName      string `yaml:"domain"`
		ServiceName     string `yaml:"service"`
		HostNameAndPort string `yaml:"host"`
	}

	registryOption struct {
		registry interface{}
		alias    string
	}

	Helper struct {
		Builder            *WorkflowClientBuilder
		Client             client.Client
		Service            workflowserviceclient.Interface
		DataConverter      encoded.DataConverter
		Logger             zap.Logger
		Config             *Configuration
		WorkerMetricScope  tally.Scope
		ServiceMetricScope tally.Scope
		Tracer             opentracing.Tracer
		CtxPropagators     []workflow.ContextPropagator
		workflowRegistries []registryOption
		activityRegistries []registryOption
	}
)

func (h *Helper) Setup(configuration *Configuration, config *utils.Config) {
	if h.Service != nil {
		return
	}

	h.Logger = *buildLogger()
	h.ServiceMetricScope = tally.NoopScope
	h.WorkerMetricScope = tally.NoopScope
	h.Config = configuration
	h.Builder = NewBuilder(&h.Logger).
		SetHostPort(h.Config.HostNameAndPort).
		SetDomain(h.Config.DomainName).
		SetDataConverter(h.DataConverter).
		SetContextPropagators(h.CtxPropagators)

	service, err := h.Builder.BuildServiceClient(config)
	if err != nil {
		panic(err)
	}

	h.Service = service
	h.Client, err = h.Builder.BuildCadenceClient(config)
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}
}

func (h *Helper) registerWorkflowAndActivity(worker worker.Worker) {
	for _, w := range h.workflowRegistries {
		if len(w.alias) == 0 {
			worker.RegisterWorkflow(w.registry)
		} else {
			worker.RegisterWorkflowWithOptions(w.registry, workflow.RegisterOptions{Name: w.alias})
		}
	}

	for _, a := range h.activityRegistries {
		if len(a.alias) == 0 {
			worker.RegisterActivity(a.registry)
		} else {
			worker.RegisterActivityWithOptions(a.registry, activity.RegisterOptions{Name: a.alias})
		}
	}
}

func (h *Helper) StartWorkers(domainName string, groupName string, options worker.Options) {
	worker := worker.New(h.Service, domainName, groupName, options)

	h.registerWorkflowAndActivity(worker)
	err := worker.Start()
	if err != nil {
		h.Logger.Error("FAILED TO START WORKER", zap.Error(err))
		panic("Failed to start worker")
	}
}

func (h *Helper) StartWorkflow(options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) *workflow.Execution {
	return h.StartWorkflowWithCtx(context.Background(), options, workflow, args...)
}

func (h *Helper) StartWorkflowWithCtx(ctx context.Context, options client.StartWorkflowOptions, workflowFunc interface{}, args ...interface{}) *workflow.Execution {
	we, err := h.Client.StartWorkflow(ctx, options, workflowFunc, args...)
	if err != nil {
		h.Logger.Error("Failed to create workflow", zap.Error(err))
		panic("Failed to create workflow.")
	} else {
		h.Logger.Info("Started Workflow", zap.String("WorkflowID", we.ID), zap.String("RunID", we.RunID))
		return we
	}
}

func (h *Helper) RegisterWorkflow(workflow any) {
	h.RegisterWorkflowWithAlias(workflow, "")
}

func (h *Helper) RegisterWorkflowWithAlias(workflow any, alias string) {
	ro := registryOption{
		registry: workflow,
		alias:    alias,
	}
	h.workflowRegistries = append(h.workflowRegistries, ro)
}

func (h *Helper) RegisterActivity(activity interface{}) {
	h.RegisterActivityWithAlias(activity, "")
}

func (h *Helper) RegisterActivityWithAlias(activity interface{}, alias string) {
	ro := registryOption{
		registry: activity,
		alias:    alias,
	}
	h.activityRegistries = append(h.activityRegistries, ro)
}
