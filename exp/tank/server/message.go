package server

type Cmd int

const (
	CQ_Move    Cmd = iota // 移动
	CQ_Attack             // 攻击
	CQ_GetRoom            // 获取整个房间
	CQ_Auth               // 登陆
)

const (
	CA_Room Cmd = iota // 获取整个房间
)

// request

type Request struct {
	Cmd  Cmd
	Body interface{}
}

// response

type Response struct {
	Cmd  Cmd
	Body interface{}
}
