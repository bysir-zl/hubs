// 分帧器
package conn_wrap

import (
	"io"
	"errors"
	"github.com/bysir-zl/hubs/core/util"
	"log"
	"encoding/binary"
)

type ProtoCoder interface {
	Read(io.Reader) ([]byte, error)       // 从一个流里读取一帧
	Write(io.Writer, []byte) (int, error) // 将一帧写入流
}

// 4个字节的头部作为长度
type LenProtoCoder struct {
}

func (p *LenProtoCoder) Read(reader io.Reader) ([]byte, error) {
	var headLen uint8
	err:= binary.Read(reader,binary.LittleEndian,headLen)
	if err != nil {
		return  nil,err
	}
	var bodyLen uint =0
	switch headLen {
	case 1:
		binary.Read(reader,binary.LittleEndian,&int8(bodyLen))
	case 2:
		binary.Read(reader,binary.LittleEndian,&int16(bodyLen))
	case 4:
		binary.Read(reader,binary.LittleEndian,&int32(bodyLen))
	case 8:
		binary.Read(reader,binary.LittleEndian,&int64(bodyLen))
	default:
		return nil,errors.New("bad head")
	}

	log.Print("bodyLen,",bodyLen)

	bs := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, bs)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (p *LenProtoCoder) Write(writer io.Writer, data []byte) (int, error) {
	bsW := make([]byte, len(data)+4)
	cmdB := util.Int322Byte(uint32(len(data)))
	bsW[0] = cmdB[0]
	bsW[1] = cmdB[1]
	bsW[2] = cmdB[2]
	bsW[3] = cmdB[3]

	copy(bsW[4:], data)
	return writer.Write(bsW)
}

func NewLenProtocal() ProtoCoder {
	return &LenProtoCoder{}
}
