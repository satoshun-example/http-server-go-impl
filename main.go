package main

import (
	"log"
	"net/http"
	"strings"
	"syscall"
)

type mux map[string]http.Handler

var defaultMux mux

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
			log.Println(string(d[before:i]))
			before = i + 1
		}
	}

	fl := strings.SplitN(string(ss[0]), " ", 3)
	method, path, proto := fl[0], fl[1], fl[2]
	log.Println(method, path, proto)

	if method != "GET" {
		_, err = syscall.Write(nfd, []byte(proto+" 405 Method Not Allowed\n"))
		return
	}

	_, err = syscall.Write(nfd, []byte(proto+" 200 OK\n"))
	if err != nil {
		panic(err)
	}

	syscall.Write(nfd, []byte("Server: original HTTP server\n\n"))
	syscall.Write(nfd, []byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<title>Test</title>
</head>
<body>Hello World!!</body>
</html>`))
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
