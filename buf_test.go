package buffpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool[int]()
	if pool == nil {
		t.Error("NewPool returned nil")
	}
}

func TestPoolInit(t *testing.T) {
	pool := NewPool[int]()
	err := pool.Init(5, 10)
	if err != nil {
		t.Errorf("Pool initialization failed: %v", err)
	}
	if pool.Available() != 5 {
		t.Errorf("Expected 5 available buffers, got %d", pool.Available())
	}
}

func TestPoolInitInvalid(t *testing.T) {
	pool := NewPool[int]()
	err := pool.Init(0, 10)
	if err == nil {
		t.Error("Expected error for invalid buffer count, got nil")
	}
	err = pool.Init(5, 0)
	if err == nil {
		t.Error("Expected error for invalid buffer size, got nil")
	}
}

func TestAcquireAndRelease(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(5, 10)

	buf, ok := pool.Acquire()
	if !ok {
		t.Error("Failed to acquire buffer")
	}
	if pool.Available() != 4 {
		t.Errorf("Expected 4 available buffers, got %d", pool.Available())
	}

	buf.Release()
	if pool.Available() != 5 {
		t.Errorf("Expected 5 available buffers after release, got %d", pool.Available())
	}
}

func TestBufferOperations(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(1, 10)

	buf, _ := pool.Acquire()
	data := buf.GetFullData()
	for i := range data {
		data[i] = i
	}
	buf.SetLength(5)

	if buf.GetLength() != 5 {
		t.Errorf("Expected length 5, got %d", buf.GetLength())
	}

	validData := buf.GetValidData()
	if len(validData) != 5 {
		t.Errorf("Expected valid data length 5, got %d", len(validData))
	}
	for i, v := range validData {
		if v != i {
			t.Errorf("Expected %d at index %d, got %d", i, i, v)
		}
	}
	buf.Release()
}

func TestConcurrentAccess(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(10, 100)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf, ok := pool.Acquire()
			if ok {
				// Simulate some work
				for i := range buf.GetFullData()[:10] {
					buf.GetFullData()[i] = i
				}
				buf.SetLength(10)
				buf.Release()
			}
		}()
	}
	wg.Wait()

	if pool.Available() != 10 {
		t.Errorf("Expected 10 available buffers after concurrent use, got %d", pool.Available())
	}
}

func TestBufferChan(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(5, 10)

	bufChan := pool.BufferChan()

	// Test acquiring buffers
	for i := 0; i < 5; i++ {
		select {
		case buf := <-bufChan:
			if buf == nil {
				t.Error("Received nil buffer from BufferChan")
			}
			if atomic.LoadInt32(&buf.inUse) != 1 {
				t.Error("Buffer from BufferChan should be marked as in use")
			}
		case <-time.After(time.Second):
			t.Error("Timed out waiting for buffer from BufferChan")
		}
	}

	// Test that BufferChan blocks when all buffers are acquired
	select {
	case <-bufChan:
		t.Error("BufferChan should block when all buffers are acquired")
	case <-time.After(time.Millisecond * 100):
		// This is the expected behavior
	}
}

func TestReset(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(5, 10)

	// Acquire all buffers
	for i := 0; i < 5; i++ {
		pool.Acquire()
	}

	if pool.Available() != 0 {
		t.Errorf("Expected 0 available buffers before reset, got %d", pool.Available())
	}

	pool.Reset()

	if pool.Available() != 5 {
		t.Errorf("Expected 5 available buffers after reset, got %d", pool.Available())
	}
}

func TestPoolRelease(t *testing.T) {
	pool := NewPool[int]()
	err := pool.Init(5, 10)
	if err != nil {
		t.Fatalf("Failed to initialize pool: %v", err)
	}

	pool.Release()

	// Test Acquire after release
	_, ok := pool.Acquire()
	if ok {
		t.Error("Acquire should fail after pool is released")
	}

	// Test BufferChan after release
	if pool.BufferChan() != nil {
		t.Error("BufferChan should return nil after pool is released")
	}

	// Test Available after release
	if pool.Available() != 0 {
		t.Error("Available should return 0 after pool is released")
	}

	// Test Reset after release
	pool.Reset() // This shouldn't panic or cause an error

	// Test double Release
	pool.Release() // This shouldn't panic or cause an error
}

func TestDoubleRelease(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(1, 10)

	buf, ok := pool.Acquire()
	if !ok {
		t.Fatal("Failed to acquire buffer")
	}

	// First Release should be fine
	buf.Release()

	// Second Release should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on double Release, but no panic occurred")
		}
	}()

	buf.Release()
}

func TestSetLengthExceedCapacity(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(1, 10)
	buf, _ := pool.Acquire()
	buf.SetLength(20)
	if buf.GetLength() != 10 {
		t.Errorf("Expected length to be capped at 10, got %d", buf.GetLength())
	}
	buf.Release()
}

func TestInitReleasedPool(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(5, 10)
	pool.Release()
	err := pool.Init(5, 10)
	if err == nil || err.Error() != "pool has been released" {
		t.Errorf("Expected 'pool has been released' error, got %v", err)
	}
}

func TestPutEdgeCases(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(1, 10)

	buf, _ := pool.Acquire()
	buf.Release()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when putting buffer to full pool")
		}
	}()
	pool.put(buf)

	buf, _ = pool.Acquire()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when putting in-use buffer")
		}
	}()
	pool.put(buf)
}

func TestAcquireEdgeCases(t *testing.T) {
	pool := NewPool[int]()
	pool.Init(1, 10)

	// 獲取唯一的 buffer
	buf1, ok := pool.Acquire()
	if !ok {
		t.Fatal("Failed to acquire first buffer")
	}

	// 嘗試獲取第二個 buffer（應該失敗）
	buf2, ok := pool.Acquire()
	if ok || buf2 != nil {
		t.Error("Expected failure when acquiring from empty pool")
	}

	// 釋放並重新獲取以覆蓋 put 和 re-acquire 的情況
	buf1.Release()
	buf2, ok = pool.Acquire()
	if !ok || buf2 == nil {
		t.Error("Failed to re-acquire buffer after release")
	}
}

func TestResetEdgeCases(t *testing.T) {
	pool := NewPool[int]()

	// 測試重置未初始化的池
	pool.Reset()
	if pool.Available() != 0 {
		t.Error("Resetting uninitialized pool should have no effect")
	}

	// 初始化並獲取所有 buffer
	pool.Init(2, 10)
	buf1, _ := pool.Acquire()
	buf2, _ := pool.Acquire()

	// 重置池（應該清除並重新填充）
	pool.Reset()

	if pool.Available() != 2 {
		t.Errorf("Expected 2 available buffers after reset, got %d", pool.Available())
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when putting buffer to full pool")
		}
	}()

	buf1.Release()
	buf2.Release()
}
