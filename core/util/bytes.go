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

