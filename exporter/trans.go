package exporter

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func convertReadOnlySpanToRequest(spans []sdktrace.ReadOnlySpan) *collectortracepb.ExportTraceServiceRequest {
	var resourceSpans []*tracepb.ResourceSpans
	// Group spans by resource
	resourceSpanMap := map[string]*tracepb.ResourceSpans{}

	for _, span := range spans {
		r := span.Resource()
		resourceKey := resourceToStringKey(r) // Grouping key

		if _, exists := resourceSpanMap[resourceKey]; !exists {
			resourceSpanMap[resourceKey] = &tracepb.ResourceSpans{
				Resource: resourceToProto(r),
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Spans: []*tracepb.Span{},
					},
				},
			}
		}

		// Append the Span to the corresponding ResourceSpans
		protoSpan := spanToProto(span)
		resourceSpanMap[resourceKey].ScopeSpans[0].Spans = append(resourceSpanMap[resourceKey].ScopeSpans[0].Spans, protoSpan)
	}

	// Collect all ResourceSpans
	for _, rs := range resourceSpanMap {
		resourceSpans = append(resourceSpans, rs)
	}

	return &collectortracepb.ExportTraceServiceRequest{
		ResourceSpans: resourceSpans,
	}
}

func resourceToStringKey(resource *resource.Resource) string {
	// Convert resource attributes to a string key for grouping
	return resource.String()
}

func resourceToProto(resource *resource.Resource) *resourcepb.Resource {
	var attributes []*commonpb.KeyValue
	for _, attr := range resource.Attributes() {
		attributes = append(attributes, &commonpb.KeyValue{
			Key: string(attr.Key),
			Value: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_StringValue{
					StringValue: attr.Value.AsString(),
				},
			},
		})
	}
	return &resourcepb.Resource{
		Attributes: attributes,
	}
}

func spanToProto(span sdktrace.ReadOnlySpan) *tracepb.Span {
	// Convert a ReadOnlySpan to tracepb.Span
	attributes := attributesToProto(span.Attributes())
	events := eventsToProto(span.Events())
	links := linksToProto(span.Links())
	traceId := span.SpanContext().TraceID()
	spanId := span.SpanContext().SpanID()
	parentSpanId := span.Parent().SpanID()
	return &tracepb.Span{
		TraceId:           traceId[:],
		SpanId:            spanId[:],
		ParentSpanId:      parentSpanId[:],
		Name:              span.Name(),
		Kind:              tracepb.Span_SpanKind(span.SpanKind()),
		StartTimeUnixNano: uint64(span.StartTime().UnixNano()),
		EndTimeUnixNano:   uint64(span.EndTime().UnixNano()),
		Attributes:        attributes,
		Events:            events,
		Links:             links,
		Status: &tracepb.Status{
			Code:    tracepb.Status_StatusCode(span.Status().Code),
			Message: span.Status().Description,
		},
	}
}

func attributesToProto(attributes []attribute.KeyValue) []*commonpb.KeyValue {
	var protoAttributes []*commonpb.KeyValue
	for _, attr := range attributes {
		protoAttributes = append(protoAttributes, &commonpb.KeyValue{
			Key: string(attr.Key),
			Value: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_StringValue{
					StringValue: attr.Value.Emit(),
				},
			},
		})
	}
	return protoAttributes
}

func eventsToProto(events []sdktrace.Event) []*tracepb.Span_Event {
	var protoEvents []*tracepb.Span_Event
	for _, event := range events {
		protoEvents = append(protoEvents, &tracepb.Span_Event{
			Name:         event.Name,
			TimeUnixNano: uint64(event.Time.UnixNano()),
			Attributes:   attributesToProto(event.Attributes),
		})
	}
	return protoEvents
}

func linksToProto(links []sdktrace.Link) []*tracepb.Span_Link {
	var protoLinks []*tracepb.Span_Link
	for _, link := range links {
		traceId := link.SpanContext.TraceID()
		spanId := link.SpanContext.SpanID()
		protoLinks = append(protoLinks, &tracepb.Span_Link{
			TraceId:    traceId[:],
			SpanId:     spanId[:],
			Attributes: attributesToProto(link.Attributes),
		})
	}
	return protoLinks
}
