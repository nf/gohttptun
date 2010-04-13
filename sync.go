package sync

import (
	"bytes"
	"io"
	"os"
)


type Buffer struct {
	readBuf chan []byte
	readCount chan int
	buf *bytes.Buffer
}

func NewBuffer(in <-chan []byte) (buf *Buffer) {
	buf = &Buffer{
		make(chan []byte),
		make(chan int),
		bytes.NewBuffer([]byte{}),
	}
	go func() {
		for {
			select {
			case b := <-buf.readBuf:
				n, _ := buf.buf.Read(b)
				buf.readCount <- n
			case b := <-in:
				buf.buf.Write(b)
			}
		}
	}()
	return
}

func (buf *Buffer) Read(b []byte) (int, os.Error) {
	buf.readBuf <- b
	return <-buf.readCount, nil
}


func ReadWriter(rw io.ReadWriter, bufSize int) (<-chan []byte, chan<- []byte) {
	in, out := make(chan []byte), make(chan []byte)
	go readLoop(rw, in, bufSize)
	go writeLoop(rw, out)
	return in, out
}

func readLoop(r io.Reader, c chan<- []byte, bufSize int) {
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

