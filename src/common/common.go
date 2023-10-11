package common

import (
	"cadtest/src/utils"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
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

type WorkflowClientBuilder struct {
	hostPort       string
	dispatcher     *yarpc.Dispatcher
	domain         string
	clientIdentity string
	metricsScope   tally.Scope
	Logger         *zap.Logger
	ctxProps       []workflow.ContextPropagator
	dataConverter  encoded.DataConverter
	tracer         opentracing.Tracer
}

func NewBuilder(logger *zap.Logger) *WorkflowClientBuilder {
	return &WorkflowClientBuilder{
		Logger: logger,
	}
}

func (b *WorkflowClientBuilder) SetHostPort(hostport string) *WorkflowClientBuilder {
	b.hostPort = hostport
	return b
}

func (b *WorkflowClientBuilder) SetDomain(domain string) *WorkflowClientBuilder {
	b.domain = domain
	return b
}

func (b *WorkflowClientBuilder) SetClientIdentity(clientIdentity string) *WorkflowClientBuilder {
	b.clientIdentity = clientIdentity
	return b
}

// SetMetricsScope sets the metrics scope for the builder
func (b *WorkflowClientBuilder) SetMetricsScope(metricsScope tally.Scope) *WorkflowClientBuilder {
	b.metricsScope = metricsScope
	return b
}

// SetDispatcher sets the dispatcher for the builder
func (b *WorkflowClientBuilder) SetDispatcher(dispatcher *yarpc.Dispatcher) *WorkflowClientBuilder {
	b.dispatcher = dispatcher
	return b
}

// SetContextPropagators sets the context propagators for the builder
func (b *WorkflowClientBuilder) SetContextPropagators(ctxProps []workflow.ContextPropagator) *WorkflowClientBuilder {
	b.ctxProps = ctxProps
	return b
}

// SetDataConverter sets the data converter for the builder
func (b *WorkflowClientBuilder) SetDataConverter(dataConverter encoded.DataConverter) *WorkflowClientBuilder {
	b.dataConverter = dataConverter
	return b
}

// SetTracer sets the tracer for the builder
func (b *WorkflowClientBuilder) SetTracer(tracer opentracing.Tracer) *WorkflowClientBuilder {
	b.tracer = tracer
	return b
}

func (b *WorkflowClientBuilder) build(config *utils.Config) error {
	if b.dispatcher != nil {
		return nil
	}
	if len(b.hostPort) == 0 {
		return errors.New("Hostport is empty.")
	}

	b.Logger.Debug("Creating RPC dispatcher outbound",
		zap.String("ServiceName", config.CadenceFrontendService),
		zap.String("HostPort", b.hostPort))

	b.dispatcher = yarpc.NewDispatcher(yarpc.Config{
		Name: config.CadenceClientService,
		Outbounds: yarpc.Outbounds{
			config.CadenceFrontendService: {Unary: grpc.NewTransport().NewSingleOutbound(b.hostPort)},
		},
	})

	if b.dispatcher != nil {
		if err := b.dispatcher.Start(); err != nil {
			b.Logger.Fatal("Failed to create outbound transport channel: %v", zap.Error(err))
		}
	}
	return nil
}

func (b *WorkflowClientBuilder) BuildCadenceClient(config *utils.Config) (client.Client, error) {
	service, err := b.BuildServiceClient(config)
	if err != nil {
		return nil, err
	}

	return client.NewClient(service, b.domain, &client.Options{
		Identity:           b.clientIdentity,
		DataConverter:      b.dataConverter,
		ContextPropagators: b.ctxProps,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}), nil
}

func (b *WorkflowClientBuilder) BuildServiceClient(config *utils.Config) (workflowserviceclient.Interface, error) {
	if err := b.build(config); err != nil {
		return nil, err
	}

	if b.dispatcher == nil {
		b.Logger.Fatal("No RPC dispatcher provided to create a connection to Cadence Service")
	}

	clientConfig := b.dispatcher.ClientConfig(config.CadenceFrontendService)
	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	), nil
}
