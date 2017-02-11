package gsproxy

import (
	"encoding/base64"
	"github.com/golang/glog"
	"net"
	"strings"
)

type Server struct {
	listenAddr            string
	base64BasicCredential string
	ln                    net.Listener
}

func NewServer(listenAddr, basicCredential string, genAuth bool) *Server {
	if genAuth {
		basicCredential = RandStringBytesMaskImprSrc(16) + ":" +
			RandStringBytesMaskImprSrc(16)
	}
	glog.Infof("use %s for auth", basicCredential)
	return &Server{listenAddr: listenAddr,
		base64BasicCredential: base64.StdEncoding.EncodeToString([]byte(basicCredential))}
}

func (s *Server) needAuth() bool {
	return s.base64BasicCredential != ""
}

// validateCredentials parse "Basic basic-credentials" and validate it
func (s *Server) validateCredentials(basicCredential string) bool {
	c := strings.Split(basicCredential, " ")
	if len(c) == 2 && strings.EqualFold(c[0], "Basic") && c[1] == s.base64BasicCredential {
		return true
	}
	return false
}

// Start a proxy
func (s *Server) Start() {
	var err error
	s.ln, err = net.Listen("tcp", s.listenAddr)
	if err != nil {
		glog.Fatalln(err)
	}

	glog.Infof("proxy listen in %s, waiting for connection...\n", s.listenAddr)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			glog.Errorln(err)
			continue
		}
		go newConn(conn, s).serve()
	}
}
