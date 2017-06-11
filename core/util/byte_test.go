package util

import (
	"testing"
)

func TestByte2Int16(t *testing.T) {
	t.Log(Byte2Int16([2]byte{255, 255}))
	t.Log(Int162Byte(256*256 - 1))

}
