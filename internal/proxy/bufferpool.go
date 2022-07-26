package proxy

import "sync"

const bufferPoolSize = 32 * 1024

type bufferPool struct {
	pool sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferPoolSize)
			},
		},
	}
}

func (b *bufferPool) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *bufferPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}
