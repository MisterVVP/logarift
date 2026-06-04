package config

import "testing"

func TestLoadUsesLocalDefaults(t *testing.T) {
	t.Setenv("LOGARIFT_API_HOST", "")
	t.Setenv("LOGARIFT_API_PORT", "")
	t.Setenv("LOGARIFT_MONGODB_URI", "")
	t.Setenv("LOGARIFT_MONGODB_DATABASE", "")
	t.Setenv("LOGARIFT_MATH_ENGINE_URL", "")
	t.Setenv("LOGARIFT_EXPORT_DIR", "")
	t.Setenv("LOGARIFT_READINESS_TIMEOUT_MS", "")
	t.Setenv("LOGARIFT_SHUTDOWN_TIMEOUT_MS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.APIPort != defaultAPIPort {
		t.Fatalf("expected default API port %s, got %s", defaultAPIPort, cfg.APIPort)
	}
	if cfg.MongoDBDatabase != defaultMongoDBDatabase {
		t.Fatalf("expected default MongoDB database %s, got %s", defaultMongoDBDatabase, cfg.MongoDBDatabase)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("LOGARIFT_API_PORT", "not-a-port")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestPublicStatusDoesNotExposeMongoURI(t *testing.T) {
	cfg := Config{
		APIHost:         "127.0.0.1",
		APIPort:         "8080",
		MongoDBURI:      "mongodb://user:secret@localhost:27017",
		MongoDBDatabase: "logarift",
		MathEngineURL:   "http://localhost:8090",
		ExportDir:       "./exports",
	}

	status := cfg.PublicStatus()
	if _, exists := status["mongodb_uri"]; exists {
		t.Fatal("PublicStatus must not expose the raw MongoDB URI")
	}
	if configured, ok := status["mongodb_uri_configured"].(bool); !ok || !configured {
		t.Fatalf("expected mongodb_uri_configured=true, got %#v", status["mongodb_uri_configured"])
	}
}
