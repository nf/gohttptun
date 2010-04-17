package main

import (
	"bytes"
	"http"
	"io"
	"net"
	"time"
)

const bufSize = 1024
func makeReadChan(r io.Reader) chan []byte {
	read := make(chan []byte)
	go func() {
		for {
			b := make([]byte, bufSize)
			n, err := r.Read(b)
			if err != nil {
				return
			}
			if n > 0 {
				read <- b[0:n]
			}
		}
	}()
	return read
}

func client(listenAddr, destAddr string) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic("Couldn't listen on "+listenAddr)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic("Couldn't accept")
	}

	// ticker to set a rate at which to hit the server
	tick := time.NewTicker(500e6)
	read := makeReadChan(conn)
	buf := bytes.NewBuffer([]byte{})
	for {
		select {
		case <-tick.C:
			// write buf to new http request
			resp, err := http.Post("http://"+destAddr+"/","application/octet-stream",buf)
			if err != nil {
				println(err.String())
				continue
			}

			// write http response response to conn
			io.Copy(conn, resp.Body)
			resp.Body.Close()
		case b := <-read:
			buf.Write(b)
		}
	}
}

func main() {
	client(":2222", "127.0.0.1:9090")
}
