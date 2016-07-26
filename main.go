package main

import (
	"bufio"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func acceptRequest() {
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

func init() {
	defaultMux = make(map[string]http.Handler)
}

func main() {
	registerHandler("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("\nHello World!!\n"))
		w.WriteHeader(200)
	})

	acceptRequest()
}
