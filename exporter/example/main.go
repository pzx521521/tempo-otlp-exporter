package main

import (
	"context"
	"github.com/pzx521521/tempo-otlp-exporter/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
	"time"
)

func initExporter() (*trace.TracerProvider, error) {
	spanExporter, err := exporter.NewMySpanExporter("tempo-prod-14-prod-ap-southeast-1.grafana.net:443")
	if err != nil {
		return nil, err
	}
	spanExporter.SetAuthInfo("1061052", "glc_xxxx")
	tp := trace.NewTracerProvider(
		trace.WithBatcher(spanExporter),
		trace.WithResource(resource.NewSchemaless(
			attribute.String("environment", "production"),
			attribute.String("service.name", "my-service"),
			attribute.String("service.version", "v1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	tp, err := initExporter()
	if err != nil {
		log.Fatalf("init Exporter fail: %v", err)
	}
	defer func() {
		_ = tp.Shutdown(context.Background())
		log.Println("TracerProvider Shutdown")
	}()
	tracer := otel.Tracer("test-Tracer")
	ctx, span := tracer.Start(context.Background(), "main-operation")
	defer span.End()
	_, span2 := tracer.Start(ctx, "sub-operation")
	time.Sleep(1 * time.Second)
	defer span2.End()
}
