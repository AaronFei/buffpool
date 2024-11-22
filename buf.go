package buffpool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Buffer[T any] struct {
	data   []T
	length int
	pool   *Pool[T]
	inUse  int32
}

func (b *Buffer[T]) SetLength(len int) {
	if len > cap(b.data) {
		len = cap(b.data)
	}
	b.length = len
}

func (b *Buffer[T]) GetLength() int {
	return b.length
}

func (b *Buffer[T]) GetValidData() []T {
	return b.data[:b.length]
}

func (b *Buffer[T]) GetFullData() []T {
	return b.data
}

func (b *Buffer[T]) Release() {
	if atomic.CompareAndSwapInt32(&b.inUse, 1, 0) {
		b.SetLength(0)
		b.pool.put(b)
	} else {
		panic("Buffer already released or was never acquired")
	}
}

type Pool[T any] struct {
	buffers       chan *Buffer[T]
	bufSize       int
	bufCount      int
	isInitialized bool
	isReleased    int32
	mu            sync.Mutex
}

func NewPool[T any]() *Pool[T] {
	return &Pool[T]{
		isInitialized: false,
		isReleased:    0,
	}
}

func (p *Pool[T]) Init(bufCount, bufSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.isReleased) == 1 {
		return fmt.Errorf("pool has been released")
	}

	if bufCount <= 0 || bufSize <= 0 {
		return fmt.Errorf("invalid buffer count or size")
	}

	p.bufCount = bufCount
	p.bufSize = bufSize
	p.buffers = make(chan *Buffer[T], bufCount)

	for i := 0; i < bufCount; i++ {
		p.buffers <- &Buffer[T]{
			data:   make([]T, bufSize),
			length: 0,
			pool:   p,
			inUse:  0,
		}
	}

	p.isInitialized = true
	return nil
}

func (p *Pool[T]) BufferChan() <-chan *Buffer[T] {
	if atomic.LoadInt32(&p.isReleased) == 1 {
		return nil
	}

	out := make(chan *Buffer[T])
	go func() {
		for buf := range p.buffers {
			if atomic.CompareAndSwapInt32(&buf.inUse, 0, 1) {
				out <- buf
			} else {
				panic("Buffer already in use")
			}
		}
		close(out)
	}()
	return out
}

func (p *Pool[T]) put(b *Buffer[T]) {
	if atomic.LoadInt32(&p.isReleased) == 1 {
		return
	}
	if atomic.LoadInt32(&b.inUse) == 0 {
		select {
		case p.buffers <- b:
			// Successfully returned to the pool
		default:
			panic("Attempting to return a buffer to a full pool")
		}
	} else {
		panic("Attempting to return a buffer that is still in use")
	}
}

func (p *Pool[T]) Acquire() (*Buffer[T], bool) {
	if atomic.LoadInt32(&p.isReleased) == 1 {
		return nil, false
	}
	select {
	case buf := <-p.buffers:
		if atomic.CompareAndSwapInt32(&buf.inUse, 0, 1) {
			return buf, true
		}
		// If the buffer is somehow already in use, put it back and try again
		p.put(buf)
		return p.Acquire()
	default:
		return nil, false
	}
}

func (p *Pool[T]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.isReleased) == 1 {
		return
	}

	if !p.isInitialized {
		return
	}

	// Clear the existing channel
	for len(p.buffers) > 0 {
		<-p.buffers
	}

	// Refill the pool
	for i := 0; i < p.bufCount; i++ {
		p.buffers <- &Buffer[T]{
			data:   make([]T, p.bufSize),
			length: 0,
			pool:   p,
			inUse:  0,
		}
	}
}

func (p *Pool[T]) Available() int {
	if atomic.LoadInt32(&p.isReleased) == 1 {
		return 0
	}
	return len(p.buffers)
}

func (p *Pool[T]) Release() {
	if atomic.CompareAndSwapInt32(&p.isReleased, 0, 1) {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Clear the channel and allow garbage collection of buffers
		for len(p.buffers) > 0 {
			<-p.buffers
		}
		close(p.buffers)

		// Reset other fields
		p.bufCount = 0
		p.bufSize = 0
		p.isInitialized = false
	}
}
