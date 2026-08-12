package redisclient

import "time"

// Options menampung seluruh konfigurasi koneksi Redis.
// Pola ini konsisten dengan functional options di httpclient.
type Options struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// EnableTLS mengaktifkan koneksi TLS (untuk Redis managed/cloud)
	EnableTLS bool
}

// Option adalah functional option untuk mengkonfigurasi Options.
type Option func(*Options)

// defaultOptions mengembalikan konfigurasi default yang aman dipakai
// kalau tidak ada option lain yang diberikan.
func defaultOptions() *Options {
	return &Options{
		Host:         "localhost",
		Port:         "6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		EnableTLS:    false,
	}
}

func WithHost(host string) Option {
	return func(o *Options) { o.Host = host }
}

func WithPort(port string) Option {
	return func(o *Options) { o.Port = port }
}

func WithPassword(password string) Option {
	return func(o *Options) { o.Password = password }
}

func WithDB(db int) Option {
	return func(o *Options) { o.DB = db }
}

func WithPoolSize(size int) Option {
	return func(o *Options) { o.PoolSize = size }
}

func WithMinIdleConns(n int) Option {
	return func(o *Options) { o.MinIdleConns = n }
}

func WithDialTimeout(d time.Duration) Option {
	return func(o *Options) { o.DialTimeout = d }
}

func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) { o.ReadTimeout = d }
}

func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) { o.WriteTimeout = d }
}

func WithTLS(enable bool) Option {
	return func(o *Options) { o.EnableTLS = enable }
}
