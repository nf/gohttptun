package main

import (
	"flag"
	"http"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"rand"
)

const (
	readTimeout = 100e6
	keyLen = 64
)

type proxy struct {
	C chan proxyPacket
	key string
	conn net.Conn
}

type proxyPacket struct {
	c *http.Conn
	r *http.Request
	done chan bool
}

func NewProxy(key, destAddr string) (p *proxy, err os.Error) {
	p = &proxy{C: make(chan proxyPacket), key: key}
	log.Stderr("Attempting connect", destAddr)
	p.conn, err = net.Dial("tcp", "", destAddr)
	if err != nil {
		return
	}
	p.conn.SetReadTimeout(readTimeout)
	log.Stderr("Connected", destAddr)
	return
}

func (p *proxy) handle(pp proxyPacket) {
	// read from the request body and write to the Conn
	_, err := io.Copy(p.conn, pp.r.Body)
	pp.r.Body.Close()
	if err == os.EOF {
		p.conn = nil
		log.Stderr("eof", p.key)
		return
	}
	// read out of the buffer and write it to conn
	pp.c.SetHeader("Content-type", "application/octet-stream")
	io.Copy(pp.c, p.conn)
	pp.done <- true
}

var queue = make(chan proxyPacket)
var createQueue = make(chan *proxy)

func handler(c *http.Conn, r *http.Request) {
	pp := proxyPacket{c, r, make(chan bool)}
	queue <- pp
	<-pp.done // wait until done before returning
}

func createHandler(c *http.Conn, r *http.Request) {
	// read destAddr
	destAddr, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(c, "Could not read destAddr",
			http.StatusInternalServerError)
		return
	}

	key := genKey()

	p, err := NewProxy(key, string(destAddr))
	if err != nil {
		http.Error(c, "Could not connect",
			http.StatusInternalServerError)
		return
	}
	createQueue <- p
	c.Write([]byte(key))
}

func proxyMuxer() {
	proxyMap := make(map[string]*proxy)
	for {
		select {
		case pp := <-queue:
			key := make([]byte, keyLen)
			// read key
			n, err := pp.r.Body.Read(key)
			if n != keyLen || err != nil {
				log.Stderr("Couldn't read key", key)
				continue
			}
			// find proxy
			p, ok := proxyMap[string(key)]
			if !ok {
				log.Stderr("Couldn't find proxy", key)
				continue
			}
			// handle
			p.handle(pp)
		case p:= <-createQueue:
			proxyMap[p.key] = p
		}
	}
}

var httpAddr = flag.String("http", ":8888", "http listen address")

func main() {
	flag.Parse()

	go proxyMuxer()

	http.HandleFunc("/", handler)
	http.HandleFunc("/create", createHandler)
	http.ListenAndServe(*httpAddr, nil)
}

func genKey() string {
	key := make([]byte, keyLen)
	for i :=0; i < keyLen; i++ {
		key[i] = byte(rand.Int())
	}
	return string(key)
}
