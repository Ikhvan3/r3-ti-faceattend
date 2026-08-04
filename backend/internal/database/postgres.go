package database

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"r3-ti-faceattend/backend/internal/config"
)

type Client struct {
	cfg  config.DatabaseConfig
	mu   sync.Mutex
	pool *pgxpool.Pool
}

func New(cfg config.DatabaseConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Ping(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pool == nil {
		if err := c.connect(ctx); err != nil {
			return err
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := c.pool.Ping(pingCtx); err != nil {
		c.closeLocked()
		return err
	}

	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closeLocked()
}

func (c *Client) connect(ctx context.Context) error {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(connectCtx, connectionString(c.cfg))
	if err != nil {
		return err
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	c.pool = pool
	return nil
}

func (c *Client) closeLocked() {
	if c.pool != nil {
		c.pool.Close()
		c.pool = nil
	}
}

func connectionString(cfg config.DatabaseConfig) string {
	values := url.Values{}
	values.Set("sslmode", cfg.SSLMode)

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host + ":" + cfg.Port,
		Path:     cfg.Name,
		RawQuery: values.Encode(),
	}

	return dsn.String()
}
