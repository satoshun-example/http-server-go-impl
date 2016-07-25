package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func writeResponse(conn net.Conn) {
	line, _ := bufio.NewReader(conn).ReadString('\n')
	s := strings.Split(line, " ")

	method, pq, proto := s[0], s[1], s[2]
	if method != "GET" {
		panic("Unknown method:" + method)
	}
	log.Println(pq, proto)

	conn.Write([]byte("HTTP/1.1 200 OK\n"))
	conn.Write([]byte("Server: original HTTP server\n"))

	// to body
	conn.Write([]byte("\n"))
	conn.Write([]byte("Hello World!!"))
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

func main() {
	acceptRequest()
}
