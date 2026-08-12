package redisclient

import (
	"errors"
	"fmt"
)

var (
	// ErrNilClient dikembalikan jika adapter dipakai sebelum di-inisialisasi
	ErrNilClient = errors.New("redisclient: client is nil, call New() first")

	// ErrKeyNotFound dikembalikan ketika key tidak ditemukan di Redis
	ErrKeyNotFound = errors.New("redisclient: key not found")

	// ErrEmptyKey dikembalikan ketika key kosong diberikan ke operasi
	ErrEmptyKey = errors.New("redisclient: key cannot be empty")

	// ErrConnectionFailed dikembalikan ketika gagal konek/ping ke Redis
	ErrConnectionFailed = errors.New("redisclient: failed to connect to redis")

	// ErrLockNotAcquired dikembalikan ketika distributed lock gagal didapat
	ErrLockNotAcquired = errors.New("redisclient: failed to acquire lock")

	// ErrLockNotOwned dikembalikan ketika mencoba unlock lock milik proses lain
	ErrLockNotOwned = errors.New("redisclient: lock is not owned by this instance")
)

// OperationError membungkus error dari operasi Redis tertentu,
// supaya caller tahu operasi apa yang gagal dan key apa yang terlibat.
type OperationError struct {
	Op  string // nama operasi, misal "GET", "SET", "DEL"
	Key string // key yang terlibat
	Err error  // error asli dari go-redis
}

func (e *OperationError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("redisclient: %s failed for key %q: %v", e.Op, e.Key, e.Err)
	}
	return fmt.Sprintf("redisclient: %s failed: %v", e.Op, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// newOpError adalah helper internal untuk membungkus error operasi
func newOpError(op, key string, err error) error {
	if err == nil {
		return nil
	}
	return &OperationError{Op: op, Key: key, Err: err}
}
