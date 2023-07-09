package mini_gin

import (
	"os"
	"time"
)

type EngineOptions struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdlTimeout        time.Duration
	Addr              string
	HandleMethodNotAllowed bool
}

// EngineOption 函数选项模式的一个优势是可以解决零值的问题。
type EngineOption func(ops *EngineOptions)

func WithReadTimeout(timeout time.Duration) EngineOption {
	return func(ops *EngineOptions) {
		ops.ReadTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) EngineOption {
	return func(ops *EngineOptions) {
		ops.ReadHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) EngineOption {
	return func(ops *EngineOptions) {
		ops.WriteTimeout = timeout
	}
}

func WithAddr(addr string) EngineOption {
	return func(ops *EngineOptions) {
		ops.Addr = addr
	}
}

func WithHandleMethodNotAllowed() EngineOption {
	return func(ops *EngineOptions) {
		ops.HandleMethodNotAllowed = true
	}
}

func (eo *EngineOptions) Apply(opts ...EngineOption) {
	for _, opt := range opts {
		opt(eo)
	}
}

func NewEngineOptions(opts ...EngineOption) *EngineOptions {
	options := &EngineOptions{
		ReadTimeout:       500 * time.Millisecond,
		ReadHeaderTimeout: 100 * time.Millisecond,
		WriteTimeout:      500 * time.Millisecond,
		IdlTimeout:        5 * time.Second,
		Addr:              getAddr(),
	}

	options.Apply(opts...)
	return options
}

func getAddr() string {
	addr := os.Getenv("MINI_GIN_SERVER_ADDR")
	if addr == "" {
		return ":8080"
	}
	return addr
}
