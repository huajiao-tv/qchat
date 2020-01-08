package gzipPool

import (
	"runtime"
	"testing"
)

var (
	Tests = []string{
		"1111111111111111111111",
		"2222222222222222222222",
		"3333333333333333333333",
		"4444444444444444444444",
		"5555555555555555555555",
		"6666666666666666666666",
		"7777777777777777777777",
		"8888888888888888888888",
		"9999999999999999999999",
		"0000000000000000000000",
	}
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestCompress(t *testing.T) {
	p := NewGzipCompressPool(1)
	d := NewGzipDecompressPool(1)
	for i := 0; i < len(Tests); i++ {
		data := p.Compress([]byte(Tests[i]))
		if string(d.Decompress(data)) != Tests[i] {
			t.Error("not eual", data, Tests[i])
		}
	}
	t.Log("succeed")
}
