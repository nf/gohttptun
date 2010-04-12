package main

import (
	"bytes"
	"./hex"
	"io"
	"net"
	"os"
)

const (
	bufSize = 1024
	hostport = "127.0.0.1:6666"
	readTimeout = 100
)

func readLoop(r io.Reader, c chan<- []byte) {
	for !closed(c) {
		b := make([]byte, bufSize)
		n, err := r.Read(b)
		if err != nil {
			return
		}
		c <- b[0:n]
	}
}

func writeLoop(w io.Writer, c <-chan []byte) {
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

func chanWrap(rw io.ReadWriter) (<-chan []byte, chan<- []byte) {
	in, out := make(chan []byte), make(chan []byte)
	go readLoop(rw, in)
	go writeLoop(rw, out)
	return in, out
}

type buffer struct {
	readBuf chan []byte
	readCount chan int
}

func NewBuffer(in <-chan []byte) (buf *buffer) {
	buf = &buffer{make(chan []byte), make(chan int)}
	bb := bytes.NewBuffer([]byte{})
	go func() {
		for {
			select {
			case b := <-buf.readBuf:
				if n, err := bb.Read(b); err == nil {
					buf.readCount <- n
				}
			case b := <-in:
				if _, err := bb.Write(b); err != nil {
					// erk
				}
			}
		}
	}()
	return
}

func (buf *buffer) Read(b []byte) (int, os.Error) {
	buf.readBuf <- b
	return <-buf.readCount, nil
}

func main() {
	conn, err := net.Dial("tcp", "", hostport)
	if err != nil {
		panic(err)
	}
	in, out := chanWrap(conn)
	for item := range hex.Encode(in) {
		out <- item
	}
}
