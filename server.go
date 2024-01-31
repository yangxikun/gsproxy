package gsproxy

import (
	"bufio"
	"encoding/base64"
	"net"
	"strings"

	"github.com/op/go-logging"
	"github.com/samber/lo"
)

var servLogger = logging.MustGetLogger("Server")

type Server struct {
	listener    net.Listener
	addr        string
	credentials []string
}

// NewServer create a proxy server
func NewServer(addr string, credentials []string, genCredential bool) *Server {
	if genCredential {
		credentials = append(credentials, RandStringBytesMaskImprSrc(16)+":"+
			RandStringBytesMaskImprSrc(16))
		servLogger.Infof("gen credentials %s for auth\n", credentials[len(credentials)-1])
	}
	for i, credential := range credentials {
		servLogger.Info(credential)
		credentials[i] = base64.StdEncoding.EncodeToString([]byte(credential))
	}
	return &Server{addr: addr, credentials: credentials}
}

// Start a proxy server
func (s *Server) Start() {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		servLogger.Fatal(err)
	}

	servLogger.Infof("proxy listen in %s, waiting for connection...\n", s.addr)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			servLogger.Error(err)
			continue
		}
		go s.newConn(conn).serve()
	}
}

// newConn create a conn to serve client request
func (s *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server: s,
		rwc:    rwc,
		brc:    bufio.NewReader(rwc),
	}
}

// isAuth return weather the client should be authenticate
func (s *Server) isAuth() bool {
	return len(s.credentials) > 0
}

// validateCredentials parse "Basic basic-credentials" and validate it
func (s *Server) validateCredential(basicCredential string) bool {
	c := strings.Split(basicCredential, " ")
	if len(c) == 2 && strings.EqualFold(c[0], "Basic") && lo.Contains(s.credentials, c[1]) {
		return true
	}
	return false
}
