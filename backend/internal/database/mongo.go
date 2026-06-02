package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
)

const defaultMongoPort = "27017"

// Client is the MVP-1 MongoDB connectivity probe. It intentionally avoids a
// repository API until MVP-2 introduces document models and collection access.
type Client struct {
	address      string
	databaseName string
	dialer       net.Dialer
}

// Connect verifies that the configured MongoDB endpoint accepts TCP
// connections. MVP-2 should replace this probe with the official MongoDB driver
// when repositories and indexes are introduced.
func Connect(ctx context.Context, cfg config.Config) (*Client, error) {
	address, err := addressFromMongoURI(cfg.MongoDBURI)
	if err != nil {
		return nil, err
	}

	client := &Client{
		address:      address,
		databaseName: cfg.MongoDBDatabase,
		dialer:       net.Dialer{},
	}
	if err := client.Ping(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// ConnectWithRetry is intentionally small and local-development oriented. It
// lets the backend wait for the Docker Compose MongoDB health check without
// adding an external dependency.
func ConnectWithRetry(ctx context.Context, cfg config.Config, interval time.Duration) (*Client, error) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}

	var lastErr error
	for {
		client, err := Connect(ctx, cfg)
		if err == nil {
			return client, nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("MongoDB unavailable before timeout: %w; last error: %v", ctx.Err(), lastErr)
		case <-time.After(interval):
		}
	}
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.address == "" {
		return fmt.Errorf("MongoDB client is not initialized")
	}
	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("connect to MongoDB at %s: %w", c.address, err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("close MongoDB readiness probe connection: %w", err)
	}
	return nil
}

func (c *Client) DatabaseName() string {
	if c == nil {
		return ""
	}
	return c.databaseName
}

func (c *Client) Close(ctx context.Context) error {
	_ = ctx
	return nil
}

func addressFromMongoURI(rawURI string) (string, error) {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return "", fmt.Errorf("parse LOGARIFT_MONGODB_URI: %w", err)
	}
	if parsed.Scheme != "mongodb" {
		return "", fmt.Errorf("LOGARIFT_MONGODB_URI must use mongodb:// scheme in MVP-1, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("LOGARIFT_MONGODB_URI must include a host")
	}

	// MVP-1 supports one local MongoDB host. Replica-set and SRV handling belong
	// with the real MongoDB driver integration in MVP-2.
	host := strings.Split(parsed.Host, ",")[0]
	if host == "" {
		return "", fmt.Errorf("LOGARIFT_MONGODB_URI must include a host")
	}
	if strings.Contains(host, ":") {
		return host, nil
	}
	return net.JoinHostPort(host, defaultMongoPort), nil
}
