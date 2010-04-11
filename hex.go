package hex

var hexDigits = []byte("0123456789ABCDEF")

func Encode(in chan []byte) (out chan []byte) {
	out = make(chan []byte)
	go func() {
		for i := range in {
			o := make([]byte, len(i)*2)
			for j, b := range i {
				o[j*2]   = hexDigits[(b&0xF0)>>4]
				o[j*2+1] = hexDigits[(b&0x0F)]
			}
			out <- o
		}
		close(out)
	}()
	return
}

func nibToByte(b byte) byte {
	for i, c := range hexDigits {
		if b == c {
			return byte(i)
		}
	}
	return 0
}

func Decode(in chan[]byte) (out chan []byte) {
	out = make(chan []byte)
	go func() {
		for i := range in {
			o := make([]byte, len(i)/2)
			for j := 0; j < len(o); j++ {
				o[j] = (nibToByte(i[j*2])<<4) + nibToByte(i[j*2+1])
			}
			out <- o
		}
		close(out)
	}()
	return
}
