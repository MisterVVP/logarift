package llmqueue

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestParseJobMessage(t *testing.T) {
	want := bson.NewObjectID()
	resp := []any{[]any{"logarift:llm_enrichment_jobs", []any{[]any{"1740000000000-0", []any{"job_id", want.Hex()}}}}}
	got, messageID, err := parseJobMessage(resp)
	if err != nil {
		t.Fatalf("parseJobMessage() error: %v", err)
	}
	if got != want || messageID != "1740000000000-0" {
		t.Fatalf("parseJobMessage() = %s, %q; want %s, %q", got.Hex(), messageID, want.Hex(), "1740000000000-0")
	}
}

func TestNewValkeyStreamRequiresValkeyURL(t *testing.T) {
	if _, err := NewValkeyStream(Options{URL: "http://localhost:6379"}); err == nil {
		t.Fatal("expected invalid scheme error")
	}
}
