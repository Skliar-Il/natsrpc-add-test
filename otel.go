package natsrpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// otelHeader is a otel trace header name.
const otelHeader = "traceparent"

// otelInject injects otel data from context into nats header.
func otelInject(ctx context.Context, callOpt *CallOptions) {
	mc := propagation.MapCarrier{}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, mc)
	if callOpt.header == nil {
		callOpt.header = map[string]string{otelHeader: mc[otelHeader]}
	} else {
		callOpt.header[otelHeader] = mc[otelHeader]
	}
	return
}

// otelExtract fetches otel data from nats header into otel header.
func otelExtract(ctx context.Context) context.Context {
	header := CallHeader(ctx)
	if header[otelHeader] != "" {
		prop := otel.GetTextMapPropagator()
		headers := make(propagation.HeaderCarrier)
		headers.Set(otelHeader, header[otelHeader])
		ctx = prop.Extract(ctx, headers)
	}
	return ctx
}
