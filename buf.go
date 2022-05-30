package simplebuf

import "fmt"

type bufferManager_t struct {
	list  chan []byte
	tmpCh chan []byte

	current []byte

	configBufCount  int
	configBufSize   int
	configSlotCount int

	isInitialized bool
}

const sizeOfDescription = 4

func GetBufferManager() bufferManager_t {
	bufferInfo := bufferManager_t{
		configSlotCount: 0,
		configBufCount:  0,
		configBufSize:   0,
		current:         nil,
		isInitialized:   false,
		tmpCh:           make(chan []byte, 1),
	}

	return bufferInfo
}

func (b *bufferManager_t) Init(slotCount int, bufCount int, bufSize int) error {
	if bufCount > slotCount || slotCount == 0 {
		return fmt.Errorf("Not have enough slot: slot count %d    buffer count %d", slotCount, bufCount)
	}

	b.configSlotCount = slotCount
	b.configBufCount = bufCount
	b.configBufSize = bufSize

	b.Reset()

	b.isInitialized = true
	return nil
}

func (b *bufferManager_t) Reset() {
	b.list = make(chan []byte, b.configSlotCount)

	for i := 0; i < b.configBufCount; i++ {
		b.Tail() <- make([]byte, sizeOfDescription+b.configBufSize)
	}
}

func (b *bufferManager_t) Next() bool {
	if !b.isInitialized {
		return false
	}

	select {
	case buf := <-b.list:
		b.current = buf
		return true
	default:
		return false
	}
}

func (b *bufferManager_t) GetCurrentDataSlice() []byte {
	if b.current != nil {
		return b.current[sizeOfDescription:]
	} else {
		return nil
	}
}

func (b *bufferManager_t) GetCurrentValidDataSlice() []byte {
	if b.current != nil {
		return b.current[sizeOfDescription : sizeOfDescription+b.GetLength()]
	} else {
		return nil
	}
}

func (b *bufferManager_t) GetCurrentSlice() []byte {
	return b.current
}

func (b *bufferManager_t) Tail() chan<- []byte {
	return b.list
}

func (b *bufferManager_t) Head() <-chan []byte {
	select {
	case b.current = <-b.list:
		b.tmpCh <- b.current
	}
	return b.tmpCh
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
	return len(b.list)
}
