package main

import (
	"bytes"
	"bufio"
	"fmt"
	"http"
	"io"
	"net"
	"time"
)

const (
	bufSize = 1024
	readTimeout = 100e6
)

func server(destAddr, httpAddr string) {
	conn, err := net.Dial("tcp", "", destAddr)
	conn.SetReadTimeout(readTimeout)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(c *http.Conn, r *http.Request) {
		// pull data from the body and copy it to the conn
		io.Copy(conn, r.Body)
		r.Body.Close()
		// read out of the buffer and write it to conn
		c.SetHeader("Content-type", "application/octet-stream")
		io.Copy(c, conn)
	})
	http.ListenAndServe(httpAddr, nil)
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

	// create a read channel
	read := make(chan []byte)
	go func() {
		for {
			b := make([]byte, bufSize)
			n, err := conn.Read(b)
			if err != nil {
				return
			}
			if n > 0 {
				read <- b[0:n]
			}
		}
	}()

	buf := bytes.NewBuffer([]byte{})
	for {
		select {
		case <-tick.C:
			// write buf to new http request
			httpConn, err := net.Dial("tcp", "", destAddr)
			if err != nil {
				println("Couldn't connect to server")
				continue
			}
			fmt.Fprintln(httpConn, "POST / HTTP/1.1")
			fmt.Fprintln(httpConn, "Content-Type: application/octet-stream")
			fmt.Fprintln(httpConn, "Content-Length: ", buf.Len())
			fmt.Fprintln(httpConn)
			buf.WriteTo(httpConn)

			// write http response response to conn
			httpBuf := bufio.NewReader(httpConn)
			resp, err := http.ReadResponse(httpBuf, "POST")
			if err != nil {
				println("Couldn't parse server response")
				continue
			}
			io.Copy(conn, resp.Body)
			resp.Body.Close()
		case b := <-read:
			buf.Write(b)
		}
	}
}

func main() {
}
