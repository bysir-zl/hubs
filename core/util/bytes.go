package util

func Byte2Int16(bs [2]byte) uint16 {
	v := uint16(bs[1]) | uint16(uint16(bs[0])<<8)
	return v
}

func Int162Byte(i uint16) (bs [2]byte) {
	bs[0] = byte(i >> 8)
	bs[1] = byte(i)
	return bs
}

func Byte4Int32(bs [4]byte) uint32 {
	v := uint32(bs[3]) |
		uint32(uint32(bs[2])<<8) |
		uint32(uint32(bs[1])<<16) |
		uint32(uint32(bs[0])<<24)
	return v
}

func Int322Byte(i uint32) (bs [4]byte) {
	bs[0] = byte(i >> 24)
	bs[1] = byte(i >> 16)
	bs[2] = byte(i >> 8)
	bs[3] = byte(i)
	return bs
}
