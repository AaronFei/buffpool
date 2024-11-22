# BuffPool

BuffPool is a generic buffer pool implementation in Go that provides efficient buffer reuse capabilities through a thread-safe pool mechanism. By reusing buffers instead of allocating new ones, it significantly reduces garbage collection pressure and improves performance in high-throughput scenarios.

## Features

- Generic implementation supporting any data type
- Thread-safe buffer acquisition and release with atomic operations
- Configurable buffer size and pool capacity
- Zero-allocation buffer reuse through efficient pooling
- Minimizes GC overhead by reusing memory buffers
- Comprehensive error handling and panic protection
- Safe pool lifecycle management (initialization, reset, and release)
- Channel-based pool implementation with non-blocking operations

## Performance Benefits

The buffer pooling mechanism provides several key performance advantages:

- **Reduced GC Pressure**: Instead of allocating and deallocating buffers repeatedly, BuffPool reuses existing buffers. This significantly reduces the number of objects that need to be garbage collected.
- **Memory Efficiency**: Pre-allocated buffers are reused efficiently, preventing memory fragmentation and reducing the total memory footprint.
- **Improved Latency**: By avoiding frequent garbage collection cycles, applications can maintain more consistent performance with lower latency spikes.
- **Better Resource Utilization**: The pool manages a fixed set of buffers, providing better control over memory usage and preventing unnecessary allocations.

## Installation

```bash
go get github.com/yourusername/buffpool
```

## Usage

### Creating and Initializing a Pool

```go
// Create a new pool for int type buffers
pool := buffpool.NewPool[int]()

// Initialize the pool with 10 buffers, each having a capacity of 1024
err := pool.Init(10, 1024)
if err != nil {
    log.Fatal(err)
}

// Don't forget to release the pool when it's no longer needed
defer pool.Release()
```

### Acquiring and Using a Buffer

```go
// Acquire a buffer from the pool
buf, ok := pool.Acquire()
if !ok {
    // Handle pool exhaustion
    return
}

// Don't forget to release the buffer when done
defer buf.Release()

// Use the buffer
data := buf.GetFullData()
// ... do something with the data

// Set the actual length of used data
buf.SetLength(100)

// Get only the valid data
validData := buf.GetValidData()
```

### Pool Management

```go
// Check available buffers
available := pool.Available()

// Reset the pool (creates new buffers)
pool.Reset()

// Get buffer channel for advanced usage
bufChan := pool.BufferChan()
```

## Advanced Usage

### Safe Concurrent Usage

```go
func producer(pool *buffpool.Pool[byte], dataChan chan<- *Buffer[byte], wg *sync.WaitGroup) {
    defer wg.Done()

    // Acquire a buffer from pool
    buf, ok := pool.Acquire()
    if !ok {
        return
    }

    // Simulate filling the buffer with data
    data := buf.GetFullData()
    // ... fill data ...
    buf.SetLength(100) // Set valid data length

    // Send buffer to consumer through channel
    dataChan <- buf
}

func consumer(dataChan <-chan *Buffer[byte], wg *sync.WaitGroup) {
    defer wg.Done()

    // Receive buffer from channel
    buf := <-dataChan

    // Process only valid data
    validData := buf.GetValidData()
    // ... process validData ...

    // Release buffer back to pool when done
    buf.Release()
}

func main() {
    pool := buffpool.NewPool[byte]()
    if err := pool.Init(10, 1024); err != nil {
        log.Fatal(err)
    }
    defer pool.Release()

    dataChan := make(chan *Buffer[byte], 10)
    var wg sync.WaitGroup

    // Start multiple producers
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go producer(pool, dataChan, &wg)
    }

    // Start multiple consumers
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go consumer(dataChan, &wg)
    }

    wg.Wait()
}
```

This example demonstrates:
- Producer acquires buffer from pool and fills it with data
- Buffer is passed to consumer through a channel
- Consumer processes only the valid data and releases the buffer
- Multiple producers and consumers can work concurrently
- Buffers are safely reused through the pool

### Using BufferChan for Stream Processing

```go
func processStream(pool *buffpool.Pool[byte]) {
    // Get a channel of buffers
    bufChan := pool.BufferChan()
    if bufChan == nil {
        // Pool has been released
        return
    }

    for buf := range bufChan {
        // Process each buffer
        // Buffer is automatically marked as in use

        // ... process data ...

        buf.Release() // Don't forget to release
    }
}
```

## Thread Safety

BuffPool implements comprehensive thread safety mechanisms:

- Atomic operations for buffer acquisition and release
- Mutex protection for pool initialization and reset
- Safe concurrent access to the buffer pool
- Protection against double-release of buffers
- Safe pool lifecycle management with release protection

## Error Handling

The library includes robust error handling:

- Pool initialization validation
- Protection against using released pools
- Panic protection for common misuse cases:
  - Releasing already released buffers
  - Returning buffers to a full pool
  - Using buffers that are still marked as in-use
- Safe cleanup through the Release method

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.
