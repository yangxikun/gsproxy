package gsproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"strings"
	"sync"
)

type conn struct {
	clientConn net.Conn
	remote     string
	serv       *Server
}

var connPool sync.Pool

func newConn(c net.Conn, serv *Server) *conn {
	if v := connPool.Get(); v != nil {
		pc := v.(*conn)
		pc.clientConn = c
		return pc
	}
	return &conn{clientConn: c, serv: serv}
}

func putConn(c *conn) {
	connPool.Put(c)
}

func (c *conn) close() {
	c.clientConn.Close()
	c.remote = ""
	putConn(c)
}

// getClientInfo parse client request header to get some information:
func (c *conn) getClientInfo() (rawHttpRequestHeader bytes.Buffer,
	host, basicCredential string, isHttps bool, err error) {

	br := newBufioReader(c.clientConn)
	tp := newTextprotoReader(br)

	defer func() {
		putBufioReader(br)
		putTextprotoReader(tp)
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()

	// First line: GET /index.html HTTP/1.0
	var s string
	if s, err = tp.ReadLine(); err != nil {
		return
	}

	method, requestURI, _, ok := parseRequestLine(s)
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

	basicCredential = mimeHeader.Get("Proxy-Authorization")

	if uriInfo.Host == "" {
		host = mimeHeader.Get("Host")
	} else {
		if strings.Index(uriInfo.Host, ":") == -1 {
			host = uriInfo.Host + ":80"
		} else {
			host = uriInfo.Host
		}
	}

	// build http request header
	rawHttpRequestHeader.WriteString(s + "\r\n")
	for k, vs := range mimeHeader {
		for _, v := range vs {
			rawHttpRequestHeader.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	rawHttpRequestHeader.WriteString("\r\n")
	return
}

// auth provide basic authentication
func (c *conn) auth(basicCredential string) bool {
	if c.serv.needAuth() == false {
		return true
	}
	if basicCredential == "" {
		glog.V(1).Infoln("ask for auth")
		// 407
		_, err := c.clientConn.Write([]byte(
			"HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"*\"\r\n\r\n"))
		if err != nil {
			glog.Errorln(err)
			return false
		}
		// get basic credential
		_, _, basicCredential, _, err = c.getClientInfo()
		if err != nil {
			glog.Errorln(err)
			return false
		}
	}
	return c.serv.validateCredentials(basicCredential)
}

// serve tunnel the client connection to remote host
func (c *conn) serve() {
	if c.clientConn == nil {
		return
	}
	defer c.close()

	rawHttpRequestHeader, host, basicCredential, isHttps, err := c.getClientInfo()
	if err != nil {
		glog.Errorln(err)
		return
	}

	if c.auth(basicCredential) == false {
		glog.Errorln("auth fail: " + basicCredential)
		return
	}

	c.remote = host
	glog.V(1).Infoln("connecting to " + c.remote)
	remote, err := net.Dial("tcp", c.remote)
	if err != nil {
		glog.Errorln(err)
		return
	}

	if isHttps {
		// if https, should sent 200 to client
		_, err = c.clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		// if not https, should sent the read header to remote
		_, err = rawHttpRequestHeader.WriteTo(remote)
	}
	if err != nil {
		glog.Errorln(err)
		return
	}

	// build bidirectional-streams
	glog.V(1).Infoln("tunneling to " + c.remote)
	go tunnel(remote, c.clientConn)
	tunnel(c.clientConn, remote)
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

type BadRequestError struct {
	what string
}

func (b *BadRequestError) Error() string {
	return b.what
}

var bufioReaderPool sync.Pool

func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	// Note: if this reader size is every changed, update
	// TestHandlerBodyClose's assumptions.
	return bufio.NewReader(r)
}

func putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bufioReaderPool.Put(br)
}

var textprotoReaderPool sync.Pool

func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
	if v := textprotoReaderPool.Get(); v != nil {
		tr := v.(*textproto.Reader)
		tr.R = br
		return tr
	}
	return textproto.NewReader(br)
}

func putTextprotoReader(r *textproto.Reader) {
	r.R = nil
	textprotoReaderPool.Put(r)
}

func tunnel(dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		glog.Errorln(err)
	}
}
