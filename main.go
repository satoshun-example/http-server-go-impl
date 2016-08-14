package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type mux map[string]http.Handler

var defaultMux mux

type addr struct {
	host string
	port string
}

type handlerFunc func(http.ResponseWriter, *http.Request)

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func (a *addr) Network() string {
	// only tcp
	return "tcp"
}

// String return a host:port
func (a *addr) String() string {
	if a == nil {
		return "<nil>"
	}
	return a.host + ":" + a.port
}

type syscallConn struct {
	sockfd int

	laddr *syscall.SockaddrInet4
	raddr *syscall.SockaddrInet4
}

func (c *syscallConn) Read(b []byte) (n int, err error) {
	n, err = syscall.Read(c.sockfd, b)
	return
}

func (c *syscallConn) Write(b []byte) (int, error) {
	syscall.Write(c.sockfd, b)
	if *debug {
		fmt.Println(string(b))
	}
	return len(b), nil
}

func (c *syscallConn) Close() (err error) {
	syscall.Shutdown(c.sockfd, syscall.SHUT_RDWR)
	syscall.Close(c.sockfd)
	return
}

func (c *syscallConn) LocalAddr() net.Addr {
	return &addr{
		host: string(c.laddr.Addr[:4]),
		port: string(c.laddr.Port),
	}
}

func (c *syscallConn) RemoteAddr() net.Addr {
	return &addr{
		host: string(c.raddr.Addr[:4]),
		port: string(c.raddr.Port),
	}
}

func (c *syscallConn) SetDeadline(t time.Time) (err error) {
	return
}

func (c *syscallConn) SetReadDeadline(t time.Time) (err error) {
	return
}

func (c *syscallConn) SetWriteDeadline(t time.Time) (err error) {
	return
}

type syscallWriter struct {
	conn    net.Conn
	code    int
	data    *bytes.Buffer
	rheader *http.Header

	req *http.Request

	contentLength int
}

func (w *syscallWriter) Header() http.Header {
	return w.req.Header
}

func (w *syscallWriter) Write(b []byte) (int, error) {
	n, err := w.data.Write(b)
	if err != nil {
		return 0, err
	}

	w.contentLength += n
	return n, nil
}

func (w *syscallWriter) WriteHeader(code int) {
	w.code = code
}

func (w *syscallWriter) dataAll() []byte {
	b, _ := ioutil.ReadAll(w.data)
	return b
}

func (w *syscallWriter) emit() {
	w.conn.Write([]byte(w.req.Proto + " " + strconv.Itoa(w.code) + " " + http.StatusText(w.code) + "\n"))

	for k, v := range *w.rheader {
		w.conn.Write([]byte(k + ": " + strings.Join(v, ",") + "\n"))
	}

	encoding := w.req.Header.Get("Accept-Encoding")
	switch encoding {
	case "gzip":
		w.conn.Write([]byte("Content-Encoding: gzip\n"))
		w.conn.Write([]byte("Content-Type: text/html; charset=UTF-8\n"))

		var b bytes.Buffer
		gz := gzip.NewWriter(&b)

		data, _ := ioutil.ReadAll(w.data)
		gz.Write(data)
		gz.Close()

		bb := b.Bytes()
		w.conn.Write([]byte([]byte("Content-Length: " + strconv.Itoa(len(bb)) + "\n\n")))
		w.conn.Write(bb)
	default: // plain
		w.conn.Write([]byte("Content-Length: " + strconv.Itoa(w.contentLength) + "\n"))
		w.conn.Write([]byte("Content-Type: text/html; charset=UTF-8\n\n"))
		w.conn.Write(w.dataAll())
	}
}

func registerHandler(pat string, handler func(http.ResponseWriter, *http.Request)) {
	defaultMux[pat] = handlerFunc(handler)
}

func acceptWithSyscall(port int) {
	// create tcp socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	laddr := &syscall.SockaddrInet4{
		Port: port,
		Addr: [4]byte{127, 0, 0, 1},
	}
	// bind port
	err = syscall.Bind(fd, laddr)

	if err != nil {
		panic(err)
	}

	// listen
	err = syscall.Listen(fd, 1)
	if err != nil {
		panic(err)
	}
	defer syscall.Shutdown(fd, syscall.SHUT_RDWR)
	defer syscall.Close(fd)

	fmt.Printf("Running on http://127.0.0.1:%d\n", port)

	// accept
	nfd, sa, err := syscall.Accept(fd)
	if err != nil {
		panic(err)
	}
	conn := &syscallConn{
		sockfd: nfd,
		laddr:  laddr,
	}
	defer conn.Close()

	raddr, ok := sa.(*syscall.SockaddrInet4)
	if !ok {
		panic("unknown protocol")
	}
	conn.raddr = raddr

	// read
	d := make([]byte, 0, 256)
	for {
		b := make([]byte, 256)
		n, err := conn.Read(b)
		if err != nil {
			panic(err)
		}
		d = append(d, b...)
		if n < 256 {
			break
		}
	}

	syscall.Shutdown(conn.sockfd, syscall.SHUT_RD)

	ss := make([][]byte, 0, 2)
	before := 0
	for i, c := range d {
		if c == '\n' {
			ss = append(ss, d[before:i])
			before = i + 1
		}
	}

	fl := strings.SplitN(string(ss[0]), " ", 3)
	method, path, proto := fl[0], fl[1], fl[2]
	proto = proto[:len(proto)-1]

	header := make(http.Header)
	for _, v := range ss[1:] {
		vv := string(v)
		vvv := strings.SplitN(vv, ":", 2)
		if len(vvv) != 2 {
			continue
		}
		header.Add(vvv[0], strings.TrimSpace(vvv[1]))
	}
	req := &http.Request{
		Proto:  proto,
		Method: method,
		Header: header,
	}
	rheader := make(http.Header)
	rheader.Add("Server", "Test HTTP server")
	writer := &syscallWriter{
		conn:    conn,
		code:    http.StatusOK,
		data:    new(bytes.Buffer),
		rheader: &rheader,
		req:     req}

	if method != "GET" {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.emit()
		return
	}

	m := false
	for pat, h := range defaultMux {
		if pat == path {
			h.ServeHTTP(writer, req)
			m = true
			break
		}
	}
	if !m {
		writer.WriteHeader(http.StatusNotFound)
	}
	writer.emit()
}

var (
	debug = flag.Bool("v", false, "verbose(debug) mode")
	port  = flag.Int("p", 8080, "port")
)

func init() {
	defaultMux = make(map[string]http.Handler)
}

func main() {
	flag.Parse()
	registerHandler("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<title>Test</title>
</head>
<body>Hello World!!</body>
</html>`))
		w.WriteHeader(200)
	})

	acceptWithSyscall(*port)
}
