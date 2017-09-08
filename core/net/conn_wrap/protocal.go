// 分帧器
package conn_wrap

import (
	"io"
	"errors"
	"encoding/binary"
)

type ProtoCoder interface {
	Read(io.Reader) ([]byte, error)       // 从一个流里读取一帧
	Write(io.Writer, []byte) (int, error) // 将一帧写入流
}

// 第一个字节是头长度1:uint8, 2:uint16, 4:uint32
// 后面x个的是内容长度
// 在后面是内容
type LenProtoCoder struct {
}

func (p *LenProtoCoder) Read(reader io.Reader) ([]byte, error) {
	var headLen uint8
	err := binary.Read(reader, binary.LittleEndian, &headLen)
	if err != nil {
		return nil, err
	}

	var bodyLen uint = 0
	switch headLen {
	case 1:
		var bodyLen8 uint8
		err = binary.Read(reader, binary.LittleEndian, &bodyLen8)
		bodyLen = uint(bodyLen8)
	case 2:
		var bodyLen16 uint16
		err = binary.Read(reader, binary.LittleEndian, &bodyLen16)
		bodyLen = uint(bodyLen16)
	case 4:
		var bodyLen32 uint32
		err = binary.Read(reader, binary.LittleEndian, &bodyLen32)
		bodyLen = uint(bodyLen32)

	default:
		return nil, errors.New("bad head")
	}
	if err != nil {
		return nil, err
	}
	bs := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (p *LenProtoCoder) Write(writer io.Writer, data []byte) (int, error) {
	bodyLen := len(data)
	var headLen uint8
	if bodyLen < 1<<(8*1) {
		headLen = 1
	} else if bodyLen < 1<<(8*2) {
		headLen = 2
	} else if bodyLen < 1<<(8*4) {
		headLen = 4
	} else {
		return 0, errors.New("too too large")
	}
	err := binary.Write(writer, binary.LittleEndian, headLen)
	if err != nil {
		return 0, err
	}
	
	switch headLen {
	case 1:
		var bodyLen8 uint8 = uint8(bodyLen)
		err = binary.Write(writer, binary.LittleEndian, &bodyLen8)
	case 2:
		var bodyLen16 uint16 = uint16(bodyLen)
		err = binary.Write(writer, binary.LittleEndian, &bodyLen16)
	case 4:
		var bodyLen32 uint32 = uint32(bodyLen)
		err = binary.Write(writer, binary.LittleEndian, &bodyLen32)
	}

	if err != nil {
		return 0, err
	}

	return writer.Write(data)
}

func NewLenProtoCoder() ProtoCoder {
	return &LenProtoCoder{}
}
