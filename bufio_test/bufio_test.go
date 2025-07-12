package bufiotest

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestAvailableBuffer(t *testing.T) {
	WriteToAvailableBuffer()
}

func WriteToAvailableBuffer() {
	w := bufio.NewWriter(os.Stdout)

	buf1 := w.AvailableBuffer()             // get b.buf[b.n:][:0] with len 0 and available cap
	buf1 = append(buf1, []byte("hello")...) // append straight to source b.buf
	w.Write(buf1)                           // copy buf1 to b.buf or (if len(b.buf)==0) write directly from b.wr (os.Stdout) to avoid copy.
	// After w.Write() w.n get the index of last copied byte

	buf2 := w.AvailableBuffer() // get slice from updated w.n index
	buf2 = append(buf2, []byte("world\n")...)
	w.Write(buf2)

	fmt.Println(string(buf1)) // len of buf1 is still 5, so we get hello

	fmt.Println(string(buf2)) // len of buf1 is 5, but it starts from updated w.n so we get world

	// Expland len to 10 by reslicing
	buf1 = buf1[:10]
	fmt.Println(string(buf1))

	w.Flush()
}

func ExampleWriter_AvailableBuffer() {
	w := bufio.NewWriter(os.Stdout)
	for _, i := range []int64{1, 2, 3, 4} {
		b := w.AvailableBuffer()
		b = strconv.AppendInt(b, i, 10)
		b = append(b, ' ')
		w.Write(b)
	}
	w.Flush()
	// Output: 1 2 3 4
}
