package main

import (
	"http"
	"io"
	"net"
	"os"
)

const readTimeout = 100e6

var conn net.Conn
var destAddr string

func connect() (err os.Error) {
	if conn != nil {
		return
	}
	conn, err = net.Dial("tcp", "", destAddr)
	if err != nil {
		println(err.String())
		return
	}
	conn.SetReadTimeout(readTimeout)
	return
}

func handler(c *http.Conn, r *http.Request) {
	if err := connect(); err != nil {
		return
	}
	// pull data from the body and copy it to the conn
	_, err := io.Copy(conn, r.Body)
	r.Body.Close()
	if err == os.EOF {
		conn = nil
		return
	}
	// read out of the buffer and write it to conn
	c.SetHeader("Content-type", "application/octet-stream")
	io.Copy(c, conn)
}

func server(dest, httpAddr string) {
	destAddr = dest
	http.HandleFunc("/", handler)
	http.ListenAndServe(httpAddr, nil)
}

func main() {
	server("10.1.1.1:22", ":9090")
}
