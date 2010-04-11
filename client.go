package main

import (
	"./hex"
	"io"
	"net"
)

const (
	bufSize = 1024
	hostport = "127.0.0.1:6666"
	readTimeout = 100
)

func readLoop(r io.Reader, c chan []byte) {
	for !closed(c) {
		b := make([]byte, bufSize)
		n, err := r.Read(b)
		if err != nil {
			return
		}
		c <- b[0:n]
	}
}

func writeLoop(w io.Writer, c chan []byte) {
	for !closed(c) {
		b := <-c
		if len(b) == 0 {
			return
		}
		_, err := w.Write(b)
		if err != nil {
			return
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "", hostport)
	if err != nil {
		panic(err)
	}
	ch := make(chan []byte)
	enc := hex.Encode(ch)
	go readLoop(conn, ch)
	writeLoop(conn, enc)

}
