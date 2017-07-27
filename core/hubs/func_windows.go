package hubs

import (
	"errors"
)

func startNewProcess(listenerFd uintptr) (pid int, err error) {
	err = errors.New("ForkExec not supported by windows ")
	return
}
