package message

import "github.com/bysir-zl/hubs/core/util"

// 定义前两位是命令
// 后面是data
type Default struct {
	cmd  uint16
	body []byte
}

func (t *Default) Marshal() (bs []byte, err error) {
	bs = make([]byte, len(t.body)+2)
	copy(bs[2:], t.body)
	cmdB := util.Int162Byte(t.cmd)
	bs[0] = cmdB[0]
	bs[1] = cmdB[1]
	return
}

func (t *Default) UnMarshal(bs []byte) (err error) {
	t.body = make([]byte, len(bs)-2)
	copy(t.body, bs[2:])
	t.cmd = util.Byte2Int16([2]byte{bs[0], bs[1]})
	return
}
