package buffers

import (
	"bytes"
	"sync"
)

// BytesBufferPool represents a pool of bytes.Buffer objects.
// BytesBufferPool 表示一个 bytes.Buffer 对象池。
type BytesBufferPool struct {
	pool       sync.Pool
	maxBufSize int
}

// NewBytesBufferPool creates a new NewBytesBufferPool with a maximum buffers size.
// NewBytesBufferPool 创建一个具有最大缓冲区大小的 NewBytesBufferPool。
func NewBytesBufferPool(maxSize int) *BytesBufferPool {
	return &BytesBufferPool{
		maxBufSize: maxSize,
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// Get retrieves a bytes.Buffer from the pool.
// Get 从池中检索一个 bytes.Buffer。
func (bp *BytesBufferPool) Get() *bytes.Buffer {
	buf := bp.pool.Get().(*bytes.Buffer)
	buf.Reset() // Reset buffers before use
	return buf
}

// Put returns a bytes.Buffer to the pool.
// If the buffers size exceeds maxBufSize, it is not put back into the pool.
// Put 将 bytes.Buffer 返回到池中。
// 如果缓冲区大小超过 maxBufSize，则不会将其放回池中。
func (bp *BytesBufferPool) Put(buf *bytes.Buffer) {
	if buf.Cap() > bp.maxBufSize {
		return // Do not put back buffers larger than maxBufSize
	}
	bp.pool.Put(buf)
}
