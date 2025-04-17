package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("lights_server")
var es *elasticsearch.Client

// InitTelemetry initializes OpenTelemetry with Elasticsearch client
func InitTelemetry() (*sdktrace.TracerProvider, error) {
	// Get Elasticsearch configuration
	apiKey := os.Getenv("ELASTIC_API_KEY")
	elasticURL := os.Getenv("ELASTIC_URL")
	if apiKey == "" || elasticURL == "" {
		return nil, fmt.Errorf("missing required environment variables: ELASTIC_API_KEY, ELASTIC_URL")
	}

	// Configure Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
		APIKey:    apiKey,
	}

	var err error
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating Elasticsearch client: %w", err)
	}

	// Test the connection
	esRes, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("testing Elasticsearch connection: %w", err)
	}
	defer esRes.Body.Close()
	log.Printf("Connected to Elasticsearch: %s", esRes.Status())

	// Create resource
	otelRes, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("lights_server"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create custom exporter that sends spans to Elasticsearch
	exporter := &elasticExporter{client: es}

	// Create TracerProvider with our custom exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(otelRes),
		sdktrace.WithBatcher(exporter),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// elasticExporter implements the SpanExporter interface
type elasticExporter struct {
	client *elasticsearch.Client
}

// ExportSpans exports spans to Elasticsearch
func (e *elasticExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		// Convert attributes to a map
		attrs := make(map[string]interface{})
		for _, attr := range span.Attributes() {
			attrs[string(attr.Key)] = attr.Value.AsInterface()
		}

		// Convert events to a slice of maps
		events := make([]map[string]interface{}, len(span.Events()))
		for i, event := range span.Events() {
			eventAttrs := make(map[string]interface{})
			for _, attr := range event.Attributes {
				eventAttrs[string(attr.Key)] = attr.Value.AsInterface()
			}
			events[i] = map[string]interface{}{
				"name":       event.Name,
				"timestamp":  event.Time.Format(time.RFC3339),
				"attributes": eventAttrs,
			}
		}

		// Create document with span data
		doc := map[string]interface{}{
			"@timestamp":    span.StartTime().Format(time.RFC3339),
			"trace_id":      span.SpanContext().TraceID().String(),
			"span_id":       span.SpanContext().SpanID().String(),
			"name":          span.Name(),
			"kind":          span.SpanKind().String(),
			"duration_ms":   span.EndTime().Sub(span.StartTime()).Milliseconds(),
			"status_code":   span.Status().Code.String(),
			"status_message": span.Status().Description,
			"attributes":    attrs,
			"events":        events,
		}

		// Encode the document
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			log.Printf("Error encoding span: %s", err)
			continue
		}

		// Index the span into Elasticsearch
		res, err := e.client.Index(
			"otel-spans",           // index name
			&buf,
			e.client.Index.WithContext(ctx),
		)
		if err != nil {
			log.Printf("Error indexing span: %s", err)
			continue
		}
		defer res.Body.Close()

		if res.IsError() {
			log.Printf("Error response from Elasticsearch: %s", res.String())
			continue
		}

		log.Printf("Logged span to Elasticsearch: %s", res.Status())
	}
	return nil
}

// Shutdown shuts down the exporter
func (e *elasticExporter) Shutdown(ctx context.Context) error {
	return nil
}

// StartSpan starts a new span with the given name and context
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

// LogEvent logs an event to the current span
func LogEvent(ctx context.Context, name string, attrs ...map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		log.Printf("No active span found for event: %s", name)
		return
	}

	// Convert attributes to span attributes
	var spanAttrs []attribute.KeyValue
	for _, attr := range attrs {
		for k, v := range attr {
			switch val := v.(type) {
			case string:
				spanAttrs = append(spanAttrs, attribute.String(k, val))
			case int:
				spanAttrs = append(spanAttrs, attribute.Int(k, val))
			case RGB:
				spanAttrs = append(spanAttrs, attribute.Int(k+".r", val.R))
				spanAttrs = append(spanAttrs, attribute.Int(k+".g", val.G))
				spanAttrs = append(spanAttrs, attribute.Int(k+".b", val.B))
			default:
				spanAttrs = append(spanAttrs, attribute.String(k, fmt.Sprintf("%v", val)))
			}
		}
	}

	span.AddEvent(name, trace.WithAttributes(spanAttrs...))
} 