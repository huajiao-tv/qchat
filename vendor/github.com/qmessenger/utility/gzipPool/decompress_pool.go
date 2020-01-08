package gzipPool

import (
	"sync"
	"compress/gzip"
	"bytes"
)

type decompress struct {
	*gzip.Reader
	buf []byte
}

type DecompressPool struct {
	l sync.Mutex
	pools chan *decompress
}

func NewGzipDecompressPool(cap int) *DecompressPool {
	return &DecompressPool{
		pools: make(chan *decompress, cap),
	}
}

// 失败后丢弃该对象
func (p *DecompressPool) Decompress(source []byte) []byte {
	var ret bytes.Buffer
	r := p.pop()
	if r.Reader == nil {
		var err error
		r.Reader, err = gzip.NewReader(bytes.NewReader(source))
		if err != nil {
			return nil
		}
	} else {
		r.Reset(bytes.NewReader(source))
	}
	for {
		n, err := r.Read(r.buf)
		if err != nil {
			return nil
		}
		if n == len(r.buf) {
			ret.Write(r.buf)
		} else if n > 0 {
			ret.Write(r.buf[0:n])
			break
		}
	}
	r.Close()
	p.push(r)
	return ret.Bytes()
}

func (p *DecompressPool) pop() *decompress {
	var r *decompress
	p.l.Lock()
	select {
	case r = <- p.pools:
	default:
		r = &decompress{
			buf: make([]byte, 512),
		}
	}
	p.l.Unlock()
	return r
}

func (p *DecompressPool) push(r *decompress) {
	p.l.Lock()
	select {
	case p.pools <- r:
	default:
	}
	p.l.Unlock()
}
