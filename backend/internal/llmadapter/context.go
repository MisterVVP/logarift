package llmadapter

import "context"

type Metadata struct {
	TraceID   string
	RequestID string
	EventID   string
	JobID     string
}

type metadataKey struct{}

func ContextWithMetadata(ctx context.Context, metadata Metadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

func metadataFromContext(ctx context.Context) Metadata {
	metadata, _ := ctx.Value(metadataKey{}).(Metadata)
	return metadata
}
