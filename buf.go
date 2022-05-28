package simplebuf

import "fmt"

type bufferManager_t struct {
	list      chan []byte
	slotCount int
	bufCount  int
	current   []byte
}

const sizeOfDescription = 4

func GetBufferManager() bufferManager_t {
	bufferInfo := bufferManager_t{
		slotCount: 0,
		bufCount:  0,
	}

	return bufferInfo
}

func (b *bufferManager_t) InitSlots(slotCount int) {
	b.list = make(chan []byte, slotCount)
	b.slotCount = slotCount
}

func (b *bufferManager_t) InitBuffers(bufferSize int, count int) error {
	if count > b.slotCount || b.slotCount == 0 {
		return fmt.Errorf("Not have enough slot: slot count %d    buffer count %d", b.slotCount, b.bufCount)
	}

	for i := 0; i < count; i++ {
		b.Tail() <- make([]byte, sizeOfDescription+bufferSize)
	}

	return nil
}

func (b *bufferManager_t) Next() bool {
	select {
	case buf := <-b.list:
		b.bufCount--
		b.current = buf
		return true
	default:
		return false
	}
}

func (b *bufferManager_t) GetCurrentDataSlice() []byte {
	return b.current[sizeOfDescription:]
}

func (b *bufferManager_t) GetCurrentSlice() []byte {
	return b.current
}

func (b *bufferManager_t) Tail() chan<- []byte {
	b.bufCount++
	return b.list
}

func (b *bufferManager_t) SetLength(len uint) {
	b.current[0] = byte(len)
	b.current[1] = byte(len >> 8)
	b.current[2] = byte(len >> 16)
	b.current[3] = byte(len >> 24)
}

func (b *bufferManager_t) GetLength() uint {
	return uint(b.current[0]) + uint(b.current[1])<<8 + uint(b.current[2])<<16 + uint(b.current[3])<<24
}

func (b *bufferManager_t) BufferCount() int {
	return b.bufCount
}
