package main

import (
	"bytes"
	"http"
	"io"
	"net"
	"time"
)

const readTimeout = 100e6

func server(destAddr, httpAddr string) {
	conn, err := net.Dial("tcp", "", destAddr)
	if err != nil {
		panic(err)
	}
	conn.SetReadTimeout(readTimeout)

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

func main() {
	server("127.0.0.1:22", ":9090")
}
