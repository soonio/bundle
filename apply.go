package bundle

import "time"

type Apply[T any] func(c *bundle[T]) *bundle[T]

// WithSize 配置打包阈值
func WithSize[T any](v int) Apply[T] {
	return func(c *bundle[T]) *bundle[T] {
		c.size = v
		return c
	}
}

// WithTimeout 配置超时时间
func WithTimeout[T any](v time.Duration) Apply[T] {
	return func(c *bundle[T]) *bundle[T] {
		c.timeout = v
		return c
	}
}

// WithPayloadSize 配置payload容量
func WithPayloadSize[T any](size int) Apply[T] {
	return func(c *bundle[T]) *bundle[T] {
		c.payloads = make(chan T, size)
		return c
	}
}
