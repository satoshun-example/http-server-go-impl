package main

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func init() {
	registerHandler("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{1, 2, 3, 4})
		w.WriteHeader(200)
	})
}

func url(port int) string {
	return "http://127.0.0.1:" + strconv.Itoa(port)
}

func TestContentLength(t *testing.T) {
	port := 8087
	go acceptWithSyscall(port)
	time.Sleep(time.Millisecond * 100)

	resp, err := http.Get(url(port))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("is not 200: actual %d", resp.StatusCode)
	}
	if resp.ContentLength != 4 {
		t.Errorf("is not 4: actual %d", resp.ContentLength)
	}
}

func TestLocalAdd(t *testing.T) {
	a := addr{
		host: "127.0.0.1",
		port: "8080",
	}
	if a.String() != "127.0.0.1:8080" {
		t.Errorf("is not 127.0.0.1:8080: actual %s", a.String())
	}
	if a.Network() != "tcp" {
		t.Errorf("is not tcp: actual %s", a.Network())
	}
}
