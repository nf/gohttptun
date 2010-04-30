package main

import (
	"bytes"
	"flag"
	"http"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const bufSize = 1024

var (
	listenAddr   = flag.String("listen", ":2222", "local listen address")
	httpAddr     = flag.String("http", "127.0.0.1:8888", "remote tunnel server")
	destAddr     = flag.String("dest", "127.0.0.1:22", "tunnel destination")
	tickInterval = flag.Int("tick", 250, "update interval (msec)")
)

// take a reader, and turn it into a channel of bufSize chunks of []byte
func makeReadChan(r io.Reader, bufSize int) chan []byte {
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

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	// initiate new session and read key
	log.Stderr("Attempting connect", *destAddr)
	buf.Write([]byte(*destAddr))
	resp, err := http.Post(
		"http://"+*httpAddr+"/create",
		"text/plain",
		buf)
	if err != nil {
		panic(err)
	}
	key, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	log.Stderr("Connected, key", key)

	// ticker to set a rate at which to hit the server
	tick := time.NewTicker(int64(*tickInterval) * 1e6)
	read := makeReadChan(conn, bufSize)
	buf.Reset()
	for {
		select {
		case <-tick.C:
			// write buf to new http request
			req := bytes.NewBuffer(key)
			buf.WriteTo(req)
			resp, err := http.Post(
				"http://"+*httpAddr+"/ping",
				"application/octet-stream",
				req)
			if err != nil {
				log.Stderr(err.String())
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
