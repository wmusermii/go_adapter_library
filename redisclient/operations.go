package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ==========================================================
// STRING OPERATIONS
// ==========================================================

// Set menyimpan value sebagai string biasa (tanpa expiry).
func (c *Client) Set(ctx context.Context, key string, value string) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Set(ctx, key, value, 0).Err(); err != nil {
		return newOpError("SET", key, err)
	}
	return nil
}

// SetWithTTL menyimpan value dengan waktu kedaluwarsa (expiry).
func (c *Client) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return newOpError("SETEX", key, err)
	}
	return nil
}

// Get mengambil value string dari key. Mengembalikan ErrKeyNotFound
// kalau key tidak ada.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", ErrNilClient
	}
	if key == "" {
		return "", ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", newOpError("GET", key, err)
	}
	return val, nil
}

// SetJSON meng-encode value apapun ke JSON lalu menyimpannya di Redis.
// Cocok untuk menyimpan struct (misal cache hasil query DB).
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	data, err := json.Marshal(value)
	if err != nil {
		return newOpError("SETJSON_MARSHAL", key, err)
	}

	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return newOpError("SETJSON", key, err)
	}
	return nil
}

// GetJSON mengambil value dari Redis dan mendekode-nya ke dalam dest
// (dest harus pointer, misal &myStruct).
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrKeyNotFound
	}
	if err != nil {
		return newOpError("GETJSON", key, err)
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return newOpError("GETJSON_UNMARSHAL", key, err)
	}
	return nil
}

// Delete menghapus satu atau lebih key sekaligus.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if len(keys) == 0 {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return newOpError("DEL", "", err)
	}
	return nil
}

// Exists mengecek apakah key ada di Redis.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if c == nil || c.rdb == nil {
		return false, ErrNilClient
	}
	if key == "" {
		return false, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, newOpError("EXISTS", key, err)
	}
	return n > 0, nil
}

// Expire mengatur TTL untuk key yang sudah ada.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		return newOpError("EXPIRE", key, err)
	}
	return nil
}

// TTL mengembalikan sisa waktu hidup sebuah key.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c == nil || c.rdb == nil {
		return 0, ErrNilClient
	}
	if key == "" {
		return 0, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return 0, newOpError("TTL", key, err)
	}
	return ttl, nil
}

// Increment menaikkan nilai integer di key sebesar 1 (cocok untuk counter, rate limiter).
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	if c == nil || c.rdb == nil {
		return 0, ErrNilClient
	}
	if key == "" {
		return 0, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, newOpError("INCR", key, err)
	}
	return val, nil
}

// IncrementBy menaikkan nilai integer di key sebesar n.
func (c *Client) IncrementBy(ctx context.Context, key string, n int64) (int64, error) {
	if c == nil || c.rdb == nil {
		return 0, ErrNilClient
	}
	if key == "" {
		return 0, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.IncrBy(ctx, key, n).Result()
	if err != nil {
		return 0, newOpError("INCRBY", key, err)
	}
	return val, nil
}

// Decrement menurunkan nilai integer di key sebesar 1.
func (c *Client) Decrement(ctx context.Context, key string) (int64, error) {
	if c == nil || c.rdb == nil {
		return 0, ErrNilClient
	}
	if key == "" {
		return 0, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.Decr(ctx, key).Result()
	if err != nil {
		return 0, newOpError("DECR", key, err)
	}
	return val, nil
}

// ==========================================================
// HASH OPERATIONS (cocok untuk menyimpan objek dengan banyak field)
// ==========================================================

// HSet menyimpan satu field di dalam hash.
func (c *Client) HSet(ctx context.Context, key string, field string, value interface{}) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.HSet(ctx, key, field, value).Err(); err != nil {
		return newOpError("HSET", key, err)
	}
	return nil
}

// HGet mengambil satu field dari hash.
func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", ErrNilClient
	}
	if key == "" {
		return "", ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", newOpError("HGET", key, err)
	}
	return val, nil
}

// HGetAll mengambil semua field+value dari sebuah hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if c == nil || c.rdb == nil {
		return nil, ErrNilClient
	}
	if key == "" {
		return nil, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, newOpError("HGETALL", key, err)
	}
	return val, nil
}

// HDelete menghapus satu atau lebih field dari hash.
func (c *Client) HDelete(ctx context.Context, key string, fields ...string) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.HDel(ctx, key, fields...).Err(); err != nil {
		return newOpError("HDEL", key, err)
	}
	return nil
}

// ==========================================================
// LIST OPERATIONS (cocok untuk queue sederhana)
// ==========================================================

// LPush menambahkan value ke bagian kiri (depan) list.
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.LPush(ctx, key, values...).Err(); err != nil {
		return newOpError("LPUSH", key, err)
	}
	return nil
}

// RPush menambahkan value ke bagian kanan (belakang) list.
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.RPush(ctx, key, values...).Err(); err != nil {
		return newOpError("RPUSH", key, err)
	}
	return nil
}

// LPop mengambil dan menghapus elemen paling kiri (depan) dari list.
// Cocok dipakai sebagai queue consumer.
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", ErrNilClient
	}
	if key == "" {
		return "", ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", newOpError("LPOP", key, err)
	}
	return val, nil
}

// LRange mengambil range elemen dari list (0 sampai -1 berarti semua elemen).
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if c == nil || c.rdb == nil {
		return nil, ErrNilClient
	}
	if key == "" {
		return nil, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	val, err := c.rdb.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, newOpError("LRANGE", key, err)
	}
	return val, nil
}

// ==========================================================
// PUB/SUB OPERATIONS
// ==========================================================

// Publish mengirim message ke sebuah channel.
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	if err := c.rdb.Publish(ctx, channel, message).Err(); err != nil {
		return newOpError("PUBLISH", channel, err)
	}
	return nil
}

// Subscribe berlangganan ke satu atau lebih channel dan mengembalikan
// *redis.PubSub yang bisa dipakai untuk menerima message via channel Go.
//
// Contoh pemakaian:
//
//	sub := client.Subscribe(ctx, "notifications")
//	defer sub.Close()
//	for msg := range sub.Channel() {
//	    fmt.Println(msg.Payload)
//	}
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Subscribe(ctx, channels...)
}

// ==========================================================
// DISTRIBUTED LOCK (berguna untuk mencegah race condition antar instance)
// ==========================================================

// AcquireLock mencoba mendapatkan lock dengan key tertentu.
// Mengembalikan token unik yang HARUS dipakai saat ReleaseLock,
// supaya hanya pemilik lock yang bisa melepasnya.
func (c *Client) AcquireLock(ctx context.Context, key string, token string, ttl time.Duration) (bool, error) {
	if c == nil || c.rdb == nil {
		return false, ErrNilClient
	}
	if key == "" {
		return false, ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	ok, err := c.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, newOpError("SETNX", key, err)
	}
	if !ok {
		return false, ErrLockNotAcquired
	}
	return true, nil
}

// releaseLockScript memastikan unlock hanya terjadi kalau token cocok
// (mencegah instance lain tidak sengaja melepas lock milik instance ini).
var releaseLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`)

// ReleaseLock melepas lock, hanya jika token yang diberikan cocok
// dengan token yang dipakai saat AcquireLock.
func (c *Client) ReleaseLock(ctx context.Context, key string, token string) error {
	if c == nil || c.rdb == nil {
		return ErrNilClient
	}
	if key == "" {
		return ErrEmptyKey
	}
	ctx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	result, err := releaseLockScript.Run(ctx, c.rdb, []string{key}, token).Int64()
	if err != nil {
		return newOpError("RELEASE_LOCK", key, err)
	}
	if result == 0 {
		return ErrLockNotOwned
	}
	return nil
}
