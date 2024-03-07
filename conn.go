package gsproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/op/go-logging"
)

var connLogger = logging.MustGetLogger("Conn")

type conn struct {
	rwc    net.Conn
	brc    *bufio.Reader
	server *Server
}

// serve tunnel the client connection to remote host
func (c *conn) serve() {
	defer c.rwc.Close()
	rawHttpRequestHeader, remote, credential, isHttps, err := c.getTunnelInfo()
	if err != nil {
		connLogger.Error(err)
		return
	}

	if c.auth(credential) == false {
		connLogger.Error("Auth fail: " + credential)
		return
	}

	connLogger.Info("connecting to " + remote)
	remoteConn, err := net.Dial("tcp", remote)
	if err != nil {
		connLogger.Error(err)
		return
	}

	if isHttps {
		// if https, should sent 200 to client
		_, err = c.rwc.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			connLogger.Error(err)
			return
		}
	} else {
		// if not https, should sent the request header to remote
		_, err = rawHttpRequestHeader.WriteTo(remoteConn)
		if err != nil {
			connLogger.Error(err)
			return
		}
	}

	// build bidirectional-streams
	connLogger.Info("begin tunnel", c.rwc.RemoteAddr(), "<->", remote)
	c.tunnel(remoteConn)
	connLogger.Info("stop tunnel", c.rwc.RemoteAddr(), "<->", remote)
}

// getClientInfo parse client request header to get some information:
func (c *conn) getTunnelInfo() (rawReqHeader bytes.Buffer, host, credential string, isHttps bool, err error) {
	tp := textproto.NewReader(c.brc)

	// First line: GET /index.html HTTP/1.0
	var requestLine string
	if requestLine, err = tp.ReadLine(); err != nil {
		return
	}

	method, requestURI, _, ok := parseRequestLine(requestLine)
	if !ok {
		err = &BadRequestError{"malformed HTTP request"}
		return
	}

	// https request
	if method == "CONNECT" {
		isHttps = true
		requestURI = "http://" + requestURI
	}

	// get remote host
	uriInfo, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return
	}

	// Subsequent lines: Key: value.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	credential = mimeHeader.Get("Proxy-Authorization")

	if uriInfo.Host == "" {
		host = mimeHeader.Get("Host")
	} else {
		if strings.Index(uriInfo.Host, ":") == -1 {
			host = uriInfo.Host + ":80"
		} else {
			host = uriInfo.Host
		}
	}

	// rebuild http request header
	rawReqHeader.WriteString(requestLine + "\r\n")
	for k, vs := range mimeHeader {
		for _, v := range vs {
			rawReqHeader.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	rawReqHeader.WriteString("\r\n")
	return
}

// auth provide basic authentication
func (c *conn) auth(credential string) bool {
	if c.server.isAuth() == false || c.server.validateCredential(credential) {
		return true
	}
	// 407
	_, err := c.rwc.Write(
		[]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"*\"\r\n\r\n"))
	if err != nil {
		connLogger.Error(err)
	}
	return false
}

// tunnel http message between client and server
func (c *conn) tunnel(remoteConn net.Conn) {
	if c.server.activeConnMetrics != nil {
		c.server.activeConnMetrics.Inc()
		defer func() {
			c.server.activeConnMetrics.Dec()
		}()
	}
	go func() {
		_, err := c.brc.WriteTo(remoteConn)
		if err != nil {
			connLogger.Warning(err)
		}
		remoteConn.Close()
	}()
	_, err := io.Copy(c.rwc, remoteConn)
	if err != nil {
		connLogger.Warning(err)
	}
}

func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

// BadRequestError 非法的请求
type BadRequestError struct {
	what string
}

func (b *BadRequestError) Error() string {
	return b.what
}
