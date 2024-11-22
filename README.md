# BuffPool

BuffPool is a generic buffer pool implementation in Go that provides efficient buffer reuse capabilities through a thread-safe pool mechanism. By reusing buffers instead of allocating new ones, it significantly reduces garbage collection pressure and improves performance in high-throughput scenarios.

## Features

- Generic implementation supporting any data type
- Thread-safe buffer acquisition and release
- Configurable buffer size and pool capacity 
- Zero-allocation buffer reuse through efficient pooling
- Minimizes GC overhead by reusing memory buffers
- Automatic buffer lifecycle management
- Channel-based pool implementation

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

### Creating a New Pool

```go
// Create a new pool for int type buffers
pool := buffpool.NewPool[int]()

// Initialize the pool with 10 buffers, each having a capacity of 1024
err := pool.Init(10, 1024)
if err != nil {
    log.Fatal(err)
}
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

// Set the actual length of used data
buf.SetLength(100)

// Get valid data slice
validData := buf.GetValidData() // Returns slice of actually used data

// Get full buffer capacity
fullData := buf.GetFullData() // Returns slice of full buffer capacity
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

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/yourusername/buffpool"
)

func main() {
    // Create and initialize pool
    pool := buffpool.NewPool[byte]()
    pool.Init(5, 1024)

    // Acquire buffer
    buf, ok := pool.Acquire()
    if !ok {
        fmt.Println("Failed to acquire buffer")
        return
    }

    // Use buffer
    data := buf.GetFullData()
    copy(data, []byte("Hello, World!"))
    buf.SetLength(13)

    // Release buffer back to pool
    buf.Release()
}
```

### Concurrent Usage

```go
func processData(pool *buffpool.Pool[int], wg *sync.WaitGroup) {
    defer wg.Done()

    buf, ok := pool.Acquire()
    if !ok {
        return
    }
    defer buf.Release()

    // Process data...
}

func main() {
    pool := buffpool.NewPool[int]()
    pool.Init(10, 1024)

    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go processData(pool, &wg)
    }
    wg.Wait()
}
```

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.
