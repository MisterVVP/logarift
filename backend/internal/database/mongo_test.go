package database

import "testing"

func TestAddressFromMongoURIUsesExplicitPort(t *testing.T) {
	address, err := addressFromMongoURI("mongodb://localhost:27018")
	if err != nil {
		t.Fatalf("addressFromMongoURI returned error: %v", err)
	}
	if address != "localhost:27018" {
		t.Fatalf("expected localhost:27018, got %s", address)
	}
}

func TestAddressFromMongoURIDefaultsPort(t *testing.T) {
	address, err := addressFromMongoURI("mongodb://mongodb")
	if err != nil {
		t.Fatalf("addressFromMongoURI returned error: %v", err)
	}
	if address != "mongodb:27017" {
		t.Fatalf("expected mongodb:27017, got %s", address)
	}
}

func TestAddressFromMongoURIRejectsUnsupportedScheme(t *testing.T) {
	_, err := addressFromMongoURI("mongodb+srv://cluster.example.com")
	if err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}
