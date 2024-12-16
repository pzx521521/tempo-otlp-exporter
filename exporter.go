package tempo_otlp_exporter

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

type TempoSpanExporter struct {
	Endpoint      string
	baseAuthToken string
	conn          *grpc.ClientConn
	client        collectortracepb.TraceServiceClient
}

func NewMySpanExporter(endpoint string) (*TempoSpanExporter, error) {
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		return nil, err
	}
	client := collectortracepb.NewTraceServiceClient(conn)
	return &TempoSpanExporter{conn: conn, client: client}, nil
}
func (t *TempoSpanExporter) SetAuthInfo(username, password string) {
	if username == "" || password == "" {
		return
	}
	auth := fmt.Sprintf("%s:%s", username, password)
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	t.baseAuthToken = "Basic " + authEncoded
}

func (t *TempoSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	headers := metadata.New(map[string]string{
		"Authorization": t.baseAuthToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, headers)
	_, err := t.client.Export(ctx, convertReadOnlySpanToRequest(spans))
	if err != nil {
		return err
	}
	slog.Info("send success", "spans_count", len(spans))
	return nil
}

func (t *TempoSpanExporter) Shutdown(_ context.Context) error {
	return t.conn.Close()
}
