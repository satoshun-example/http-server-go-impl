package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
)

type mux map[string]http.Handler

var defaultMux mux

type writer struct {
	conn net.Conn
	code int
	b    [][]byte
}

func (w *writer) Header() http.Header {
	return nil
}

func (w *writer) Write(b []byte) (int, error) {
	w.b = append(w.b, b)
	return len(b), nil
}

func (w *writer) WriteHeader(code int) {
	w.code = code
}

func (w *writer) emit() {
	// TODO: OK?
	w.conn.Write([]byte("HTTP/1.1 " + strconv.Itoa(w.code) + " OK\n"))
	w.conn.Write([]byte("Server: original HTTP server\n"))

	for i := range w.b {
		w.conn.Write(w.b[i])
	}
}

func writeResponse(conn net.Conn) {
	line, _ := bufio.NewReader(conn).ReadString('\n')
	s := strings.Split(line, " ")

	method, pq, proto := s[0], s[1], s[2]
	if method != "GET" {
		panic("Unknown method:" + method)
	}

	req := &http.Request{
		Method: method,
		Proto:  proto,
	}

	for pat, m := range defaultMux {
		if pat == pq {
			w := &writer{conn: conn, code: 200, b: make([][]byte, 0, 1)}
			m.ServeHTTP(w, req)
			w.emit()
			break
		}
	}
}

func registerHandler(pat string, handler func(http.ResponseWriter, *http.Request)) {
	defaultMux[pat] = http.HandlerFunc(handler)
}

func acceptWithNetListen() {
	// TODO: does not use net#Listen
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}

	// TODO: use go keyword
	writeResponse(conn)
	conn.Close()
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
			log.Println(string(d[before:i]))
			before = i + 1
		}
	}

	fl := strings.Split(string(ss[0]), " ")
	method, path, proto := fl[0], fl[1], fl[2]
	log.Println(method, path, proto)

	_, err = syscall.Write(nfd, []byte("HTTP/1.1 200 OK\n"))
	if err != nil {
		panic(err)
	}

	syscall.Write(nfd, []byte("Server: original HTTP server\n\n"))
	syscall.Write(nfd, []byte("Hello World!!"))
}

func init() {
	defaultMux = make(map[string]http.Handler)
}

func main() {
	registerHandler("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("\nHello World!!\n"))
		w.WriteHeader(200)
	})

	acceptWithSyscall()
}
