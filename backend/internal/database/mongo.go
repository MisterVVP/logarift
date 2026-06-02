package database

import (
	"context"
	"fmt"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
	name     string
}

func Connect(ctx context.Context, cfg config.Config) (*Client, error) {
	opts := options.Client().ApplyURI(cfg.MongoDBURI)
	mc, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect MongoDB: %w", err)
	}
	c := &Client{client: mc, database: mc.Database(cfg.MongoDBDatabase), name: cfg.MongoDBDatabase}
	if err := c.Ping(ctx); err != nil {
		_ = mc.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}
	return c, nil
}
func ConnectWithRetry(ctx context.Context, cfg config.Config, interval time.Duration) (*Client, error) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	var lastErr error
	for {
		c, err := Connect(ctx, cfg)
		if err == nil {
			return c, nil
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
	if c == nil || c.client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}
	return c.client.Ping(ctx, readpref.Primary())
}
func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Disconnect(ctx)
}
func (c *Client) Database() *mongo.Database { return c.database }
func (c *Client) DatabaseName() string {
	if c == nil {
		return ""
	}
	return c.name
}
