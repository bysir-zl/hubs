// 分帧器
package channel

import (
	"io"
	"encoding/binary"
)

type ProtoCol interface {
	Read(io.Reader) ([]byte, error)       // 从一个流里读取一帧
	Write(io.Writer, []byte) (int, error) // 将一帧写入流
}

// 第一个字节是头长度1:uint8, 2:uint16, 4:uint32
// 后面x个的是内容长度
// 在后面是内容
type LenProtoCol struct {
}

func (p *LenProtoCol) Read(reader io.Reader) ([]byte, error) {
	var bodyLen32 uint32
	err := binary.Read(reader, binary.LittleEndian, &bodyLen32)
	if err != nil {
		return nil, err
	}

	bs := make([]byte, bodyLen32)
	_, err = io.ReadFull(reader, bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (p *LenProtoCol) Write(writer io.Writer, data []byte) (int, error) {
	bodyLen32 := uint32(len(data))

	err := binary.Write(writer, binary.LittleEndian, &bodyLen32)
	if err != nil {
		return 0, err
	}

	i, e := writer.Write(data)
	return i, e
}

func NewLenProtoCol() ProtoCol {
	return &LenProtoCol{}
}
