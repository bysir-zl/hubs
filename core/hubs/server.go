package hubs

import (
	"github.com/bysir-zl/hubs/core/net/listener"
	"github.com/bysir-zl/hubs/core/net/conn_wrap"
	"sync"
	"os"
	"github.com/bysir-zl/hubs/core/util"
)

type ConnHandler interface {
	Server(*Server, conn_wrap.Interface)
}

type Server struct {
	addr       string
	listener   listener.Interface
	isGraceful bool
	handler    ConnHandler
	wg         sync.WaitGroup
	*Manager
}

const GRACEFUL_ENVIRON_STRING = "GRACEFUL_ENVIRON_STRING"
const GRACEFUL_ENVIRON_KEY = "GRACEFUL_ENVIRON_KEY"

func (s *Server) Run() (err error) {
	err = s.listener.Listen(s.addr, s.isGraceful)
	if err != nil {
		return
	}

	for {
		c, e := s.listener.Accept()
		if e != nil {
			err = e
			if listener.Err_Stoped != e {
				util.Logger.Warn("accept err ", e)
			}
			break
		}
		s.wg.Add(1)
		go func() {
			s.handler.Server(s, c)
			s.wg.Done()
		}()
	}

	// 如果accept错误,则listener已关闭, 等到已经连接的连接关闭
	s.wg.Wait()

	return
}

func (s *Server) Stop() {
	err := s.listener.Close()
	if err != nil {
		util.Logger.Warn(err)
	} else {
		util.Logger.Info("server is stopping")
	}
}

func (s *Server) Graceful() {
	fd, err := s.listener.Fd()
	if err != nil {
		util.Logger.Error("Graceful err ", err)
		return
	}
	pid, err := startNewProcess(fd)
	if err != nil {
		util.Logger.Error("Graceful err ", err)
		return
	}
	util.Logger.Info("ForkExec success,pid: ", pid)

	s.Stop()
}

func New(addr string, listener listener.Interface, h ConnHandler) *Server {
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}
	return &Server{
		isGraceful: isGraceful,
		listener:   listener,
		handler:    h,
		addr:       addr,
		Manager:    NewManager(),
	}
}
