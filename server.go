package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()

	}

}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	fmt.Println("connect success")
	user := NewUser(conn, s)
	user.Online()

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("read buf error", err)
				return
			}
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isLive <- true

		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Minute * 10):
			user.SendMsg("you have been removed\n")
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net not listen", err)
		return
	}
	defer listener.Close()
	go s.ListenMessager()
	for {
		accept, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error", err)
			continue
		}

		go s.Handler(accept)

	}
}
