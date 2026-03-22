package service

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// NoopTracerProvider is a simple no-op tracer provider
type NoopTracerProvider struct{}

func (p NoopTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return trace.NewNoopTracerProvider().Tracer(name)
}

// TracingService provides distributed tracing capabilities
type TracingService struct {
	tracer    trace.Tracer
	provider  *sdktrace.TracerProvider
	serviceName string
}

// TracingConfig for tracing configuration
type TracingConfig struct {
	ServiceName    string
	Environment    string
	Version        string
	OTLPEndpoint   string
	SamplingRate   float64
	Enabled        bool
}

// NewTracingService creates a new tracing service
func NewTracingService(cfg *TracingConfig) (*TracingService, error) {
	if !cfg.Enabled {
		return &TracingService{
			tracer: trace.NewNoopTracerProvider().Tracer("noop"),
		}, nil
	}

	// Create OTLP exporter
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
	)

	// Set global provider
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracingService{
		tracer:      provider.Tracer(cfg.ServiceName),
		provider:    provider,
		serviceName: cfg.ServiceName,
	}, nil
}

// Shutdown shuts down the tracing service
func (t *TracingService) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (t *TracingService) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// TraceTask traces a task execution
func (t *TracingService) TraceTask(ctx context.Context, taskID, skillID string) (context.Context, *TaskSpan) {
	ctx, span := t.tracer.Start(ctx, "task.execute",
		trace.WithAttributes(
			attribute.String("task.id", taskID),
			attribute.String("task.skill_id", skillID),
		),
	)

	return ctx, &TaskSpan{
		span:    span,
		start:   time.Now(),
		taskID:  taskID,
		skillID: skillID,
	}
}

// TaskSpan represents a traced task
type TaskSpan struct {
	span    trace.Span
	start   time.Time
	taskID  string
	skillID string
}

// SetStatus sets the span status
func (s *TaskSpan) SetStatus(status string) {
	s.span.SetAttributes(attribute.String("task.status", status))
}

// SetProgress sets the task progress
func (s *TaskSpan) SetProgress(current, total int) {
	s.span.SetAttributes(
		attribute.Int("task.progress.current", current),
		attribute.Int("task.progress.total", total),
	)
}

// SetError marks the span as errored
func (s *TaskSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// End ends the span
func (s *TaskSpan) End() {
	duration := time.Since(s.start)
	s.span.SetAttributes(
		attribute.Int64("task.duration_ms", duration.Milliseconds()),
	)
	s.span.End()
}

// TracePipeline traces a pipeline execution
func (t *TracingService) TracePipeline(ctx context.Context, pipelineID, mode string) (context.Context, *PipelineSpan) {
	ctx, span := t.tracer.Start(ctx, "pipeline.execute",
		trace.WithAttributes(
			attribute.String("pipeline.id", pipelineID),
			attribute.String("pipeline.mode", mode),
		),
	)

	return ctx, &PipelineSpan{
		span:       span,
		start:      time.Now(),
		pipelineID: pipelineID,
		mode:       mode,
	}
}

// PipelineSpan represents a traced pipeline
type PipelineSpan struct {
	span       trace.Span
	start      time.Time
	pipelineID string
	mode       string
	stepCount  int
}

// AddStep adds a step event
func (s *PipelineSpan) AddStep(stepID, status string, duration time.Duration) {
	s.stepCount++
	s.span.AddEvent("step.completed",
		trace.WithAttributes(
			attribute.String("step.id", stepID),
			attribute.String("step.status", status),
			attribute.Int64("step.duration_ms", duration.Milliseconds()),
		),
	)
}

// SetStatus sets the span status
func (s *PipelineSpan) SetStatus(status string) {
	s.span.SetAttributes(attribute.String("pipeline.status", status))
}

// SetError marks the span as errored
func (s *PipelineSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// End ends the span
func (s *PipelineSpan) End() {
	duration := time.Since(s.start)
	s.span.SetAttributes(
		attribute.Int64("pipeline.duration_ms", duration.Milliseconds()),
		attribute.Int("pipeline.step_count", s.stepCount),
	)
	s.span.End()
}

// TraceStep traces a step execution
func (t *TracingService) TraceStep(ctx context.Context, stepID, cli string) (context.Context, *StepSpan) {
	ctx, span := t.tracer.Start(ctx, "step.execute",
		trace.WithAttributes(
			attribute.String("step.id", stepID),
			attribute.String("step.cli", cli),
		),
	)

	return ctx, &StepSpan{
		span:   span,
		start:  time.Now(),
		stepID: stepID,
		cli:    cli,
	}
}

// StepSpan represents a traced step
type StepSpan struct {
	span   trace.Span
	start  time.Time
	stepID string
	cli    string
}

// SetAction sets the step action
func (s *StepSpan) SetAction(action string) {
	s.span.SetAttributes(attribute.String("step.action", action))
}

// SetStatus sets the span status
func (s *StepSpan) SetStatus(status string) {
	s.span.SetAttributes(attribute.String("step.status", status))
}

// SetError marks the span as errored
func (s *StepSpan) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// End ends the span
func (s *StepSpan) End() {
	duration := time.Since(s.start)
	s.span.SetAttributes(
		attribute.Int64("step.duration_ms", duration.Milliseconds()),
	)
	s.span.End()
}

// SpanFromContext returns the current span from context
func (t *TracingService) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span
func (t *TracingService) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetAttributes sets attributes on the current span
func (t *TracingService) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}