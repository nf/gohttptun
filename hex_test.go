package hex

import "testing"

func TestEncodeDecode(t *testing.T) {
	str := "This is a test string."

	in := make(chan []byte)
	out := Encode(in)
	in <- []byte(str)
	b := <-out

	in2 := make(chan []byte)
	out2 := Decode(in2)
	in2 <- b

	outstr := string(<-out2)
	if outstr != str {
		t.Error("got '%s' expected '%s'.", outstr, str)
	}
}
