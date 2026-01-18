package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
	"strconv"
	"sync/atomic"
)

type ServerAddr struct {
	networkType string
	networkAddr string
}

func (sAddr ServerAddr) Network() string {
	return sAddr.networkType
}

func (sAddr ServerAddr) String() string {
	return sAddr.networkAddr
}

type Server struct {
	state *atomic.Bool
	serverAddr ServerAddr
	listener net.Listener
	handler Handler
}

func newServer(port int, handlerFunc Handler) *Server {
	fmt.Printf("Creating new server on port%d\n", port)
	var serverState atomic.Bool
	serverState.Store(false)
	return &Server{
		state: &serverState,
		serverAddr: ServerAddr{
			networkType: "tcp",
			networkAddr: ":"+strconv.Itoa(port),
		},
		listener: nil,
		handler: handlerFunc,
	}
}

func Serve(port int, handlerFunc Handler) (*Server, error) {
	server := newServer(port, handlerFunc)
	go server.listen()
	return server, nil
}

func (s *Server) Accept() (net.Conn, error) {
	conn, err := s.listener.Accept()

	return conn, err
}

func (s *Server) Addr() (net.Addr) {
	return s.serverAddr
}

func (s *Server) Close() error {
	old := s.state.Swap(false)
	var err error
	if !old {
		err = fmt.Errorf("Server is already closed")
	}
	return err
}

func (s *Server) listen() {
	s.state.Store(true)
	
	listener, err := net.Listen(s.serverAddr.networkType, s.serverAddr.networkAddr)
	if err != nil {
		s.state.Store(false)
		s.Close()
		return
	}
	s.listener = listener

	for s.state.Load() {
		conn, err := s.Accept()
		if err != nil {
			fmt.Println("Failed to read connection. Don't know how to respond" + err.Error())
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	writer := response.NewWriter()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteResponse(400, err.Error())
	} else {
		s.handler(&writer, req);
	}

	conn.Write([]byte(writer.ReadBuffer()))
}