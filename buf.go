package simplebuf

const sizeOfDescription = 4

func InitBufferList(bufferSize int, bufferCount int) [][]byte {
	bufferList := [][]byte{}
	for i := 0; i < bufferCount; i++ {
		bufferList = append(bufferList, make([]byte, sizeOfDescription+bufferSize))
	}

	return bufferList
}

func SetLength(buf []byte, len uint) {
	buf[0] = byte(len)
	buf[1] = byte(len >> 8)
	buf[2] = byte(len >> 16)
	buf[3] = byte(len >> 24)
}

func GetLength(buf []byte) uint {
	return uint(buf[0]) + uint(buf[1])<<8 + uint(buf[2])<<16 + uint(buf[3])<<24
}

func GetDataSlice(buf []byte) []byte {
	return buf[sizeOfDescription:]
}
