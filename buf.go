package simplebuf

type BufferInfo_t struct {
	list    chan []byte
	current []byte
}

const sizeOfDescription = 4

func InitBufferList(bufferSize int, bufferCount int) BufferInfo_t {
	bufferInfo := BufferInfo_t{
		list: make(chan []byte, bufferCount),
	}

	for i := 0; i < bufferCount; i++ {
		bufferInfo.list <- make([]byte, sizeOfDescription+bufferSize)
	}

	return bufferInfo
}

func (b *BufferInfo_t) Next() []byte {
	select {
	case buf := <-b.list:
		b.current = buf
		return buf[sizeOfDescription:]
	default:
		return nil
	}
}

func (b *BufferInfo_t) Tail(buf []byte) {
	b.list <- buf
}

func (b *BufferInfo_t) SetLength(len uint) {
	b.current[0] = byte(len)
	b.current[1] = byte(len >> 8)
	b.current[2] = byte(len >> 16)
	b.current[3] = byte(len >> 24)
}

func (b *BufferInfo_t) GetLength() uint {
	return uint(b.current[0]) + uint(b.current[1])<<8 + uint(b.current[2])<<16 + uint(b.current[3])<<24
}

func GetLength(buf []byte) uint {
	return uint(buf[0]) + uint(buf[1])<<8 + uint(buf[2])<<16 + uint(buf[3])<<24
}

func GetDataSlice(buf []byte) []byte {
	return buf[sizeOfDescription:]
}
