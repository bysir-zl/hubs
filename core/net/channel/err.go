package channel

import "errors"

var (
	Err_CloseByPong = errors.New("long time no pong")
	Err_CloseByPing = errors.New("long time no ping")
	Err_CloseDefault = errors.New("closed")
)
