package main

import (
	"net"
	"./sync"
	"time"
)

const (
	bufSize = 1024
	hostport = "127.0.0.1:6666"
	readTimeout = 100
)

func main() {
	conn, err := net.Dial("tcp", "", hostport)
	if err != nil {
		panic(err)
	}
	in, out := sync.ReadWriter(conn, bufSize)
	buf := sync.NewBuffer(in)
	for {
		time.Sleep(2e9)
		b := make([]byte, bufSize)
		n, _ := buf.Read(b)
		if n > 0 {
			out <- b[0:n]
		}
	}
}
