package main

import (
	"net/http"
	"strconv"
	"strings"
	"syscall"
)

type mux map[string]http.Handler

var defaultMux mux

type syscallWriter struct {
	socketfd int
	code     int
	data     [][]byte
	req      *http.Request
}

func (w *syscallWriter) Header() http.Header {
	return w.req.Header
}

func (w *syscallWriter) Write(b []byte) (int, error) {
	w.data = append(w.data, b)
	return len(b), nil
}

func (w *syscallWriter) WriteHeader(code int) {
	w.code = code
}

func (w *syscallWriter) emit() {
	syscall.Write(w.socketfd, []byte(w.req.Proto+" "+strconv.Itoa(w.code)+" "+http.StatusText(w.code)+"\n"))

	for k, v := range w.req.Header {
		syscall.Write(w.socketfd, []byte(k+": "+strings.Join(v, ",")+"\n"))
	}

	syscall.Write(w.socketfd, []byte{'\n'})

	for i := range w.data {
		syscall.Write(w.socketfd, w.data[i])
	}
}

func registerHandler(pat string, handler func(http.ResponseWriter, *http.Request)) {
	defaultMux[pat] = http.HandlerFunc(handler)
}

func acceptWithSyscall() {
	// create tcp socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	// bind port
	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: 8080,
		Addr: [4]byte{127, 0, 0, 1},
	})
	if err != nil {
		panic(err)
	}

	// listen
	err = syscall.Listen(fd, 1)
	if err != nil {
		panic(err)
	}

	// accept
	nfd, _, err := syscall.Accept(fd)
	if err != nil {
		syscall.Close(fd)
		panic(err)
	}
	defer syscall.Close(nfd)

	// read
	d := make([]byte, 0, 256)
	for {
		b := make([]byte, 256)
		n, err := syscall.Read(nfd, b)
		if err != nil {
			panic(err)
		}
		d = append(d, b...)
		if n < 256 {
			break
		}
	}

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
	header.Add("Server", "Test HTTP server")
	req := &http.Request{
		Proto:  proto,
		Method: method,
		Header: header,
	}
	writer := &syscallWriter{
		socketfd: nfd,
		code:     http.StatusOK,
		data:     make([][]byte, 0, 2),
		req:      req}

	if method != "GET" {
		writer.WriteHeader(405)
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
		writer.WriteHeader(404)
	}
	writer.emit()
}

func init() {
	defaultMux = make(map[string]http.Handler)
}

func main() {
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

	acceptWithSyscall()
}
