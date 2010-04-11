package main

import (
	"fmt"
	"./hex"
)

func main() {
	in := make(chan []byte)
	out := hex.Encode(in)
	in <- []byte("Andrew Is A Programmer")
	b := <-out
	fmt.Println(string(b))

	in2 := make(chan []byte)
	out2 := hex.Decode(in2)
	in2 <- b
	fmt.Println(string(<-out2))
}
