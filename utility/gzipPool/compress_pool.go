package gzipPool

import (
	"bytes"
	"compress/gzip"
	"sync"
)

type CompressPool struct {
	l     sync.Mutex
	pools chan *gzip.Writer
}

func NewGzipCompressPool(cap int) *CompressPool {
	return &CompressPool{
		pools: make(chan *gzip.Writer, cap),
	}
}

// 失败后丢弃该对象
func (p *CompressPool) Compress(source []byte) []byte {
	var buf bytes.Buffer
	w := p.pop()
	w.Reset(&buf)
	_, err := w.Write(source)
	if err != nil {
		return nil
	}
	err = w.Close()
	if err != nil {
		return nil
	}
	ret := buf.Bytes()
	p.push(w)
	return ret
}

func (p *CompressPool) pop() *gzip.Writer {
	var w *gzip.Writer
	p.l.Lock()
	select {
	case w = <-p.pools:
	default:
		w = gzip.NewWriter(new(bytes.Buffer))
	}
	p.l.Unlock()
	return w
}

func (p *CompressPool) push(w *gzip.Writer) {
	p.l.Lock()
	select {
	case p.pools <- w:
	default:
	}
	p.l.Unlock()
}
