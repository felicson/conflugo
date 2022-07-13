package buf

import "bytes"

type Buffer struct {
	bytes.Buffer
}

func (mb *Buffer) Write(b []byte) (int, error) {
	b = bytes.ReplaceAll(b, []byte(`\\_`), []byte("_"))
	b = bytes.ReplaceAll(b, []byte("&"), []byte("&amp"))
	return mb.Buffer.Write(b)
}
