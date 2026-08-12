package redisclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client adalah adapter utama untuk Redis, membungkus *redis.Client
// dari go-redis supaya pemakaian di seluruh project konsisten
// dan mudah diganti implementasinya kalau perlu.
type Client struct {
	rdb     *redis.Client
	options *Options
}

// New membuat instance Client baru dan langsung melakukan Ping
// untuk memastikan koneksi ke Redis berhasil.
//
// Contoh pemakaian:
//
//	client, err := redisclient.New(
//	    redisclient.WithHost("localhost"),
//	    redisclient.WithPort("6379"),
//	    redisclient.WithPassword("redispassword123"),
//	    redisclient.WithDB(0),
//	)
func New(opts ...Option) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	redisOpts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", options.Host, options.Port),
		Password:     options.Password,
		DB:           options.DB,
		PoolSize:     options.PoolSize,
		MinIdleConns: options.MinIdleConns,
		DialTimeout:  options.DialTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	}

	if options.EnableTLS {
		redisOpts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	rdb := redis.NewClient(redisOpts)

	ctx, cancel := context.WithTimeout(context.Background(), options.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	return &Client{
		rdb:     rdb,
		options: options,
	}, nil
}

// Ping mengecek apakah koneksi ke Redis masih hidup.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if err := c.rdb.Ping(ctx).Err(); err != nil {
		return newOpError("PING", "", err)
	}
	return nil
}

// Close menutup seluruh koneksi pool ke Redis.
// Panggil ini saat aplikasi shutdown (biasanya via defer di container.go).
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// Raw mengembalikan instance *redis.Client asli dari go-redis,
// untuk kasus di mana kamu butuh fitur yang belum di-wrap di adapter ini.
func (c *Client) Raw() *redis.Client {
	if c == nil {
		return nil
	}
	return c.rdb
}

// Stats mengembalikan statistik pool koneksi (berguna untuk monitoring/health check).
func (c *Client) Stats() *redis.PoolStats {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.PoolStats()
}

// contextWithTimeout adalah helper internal untuk membuat context
// dengan default timeout kalau caller tidak memberi context ber-deadline.
func (c *Client) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, 5*time.Second)
}
