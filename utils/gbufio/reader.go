package gbufio

import (
	"bufio"
	"bytes"
	"io"
)

type Reader struct {
	err error
	buf []byte
	
	rd   io.Reader
	rpos int
	wpos int

	slice sliceAlloc
}

type sliceAlloc struct {
	buf []byte
}

func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, 1024)
}

func NewReaderSize(rd io.Reader, size int) *Reader {
	if size <= 0 {
		size = 1024
	}
	return &Reader{rd: rd, buf: make([]byte, size)}
}

func (b *Reader) fill() error {
	if b.err != nil {
		return b.err
	}
	if b.rpos > 0 {
		n := copy(b.buf, b.buf[b.rpos:b.wpos])
		b.rpos = 0
		b.wpos = n
	}
	n, err := b.rd.Read(b.buf[b.wpos:])
	if err != nil {
		b.err = err
	} else if n == 0 {
		b.err = io.ErrNoProgress
	} else {
		b.wpos += n
	}
	return b.err
}

func (b *Reader) buffered() int {
	return b.wpos - b.rpos
}
func (b *Reader) ReadByte() (byte, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.buffered() == 0 {
		if b.fill() != nil {
			return 0, b.err
		}
	}
	c := b.buf[b.rpos]
	b.rpos++
	return c, nil
}

func (b *Reader) ReadBytes(delim byte) ([]byte, error) {
	var full [][]byte
	var last []byte
	var size int
	for last == nil {
		f, err := b.ReadSlice(delim)
		if err != nil {
			if err != bufio.ErrBufferFull {
				return nil, b.err
			}
			dup := b.slice.Make(len(f))
			copy(dup, f)
			full = append(full, dup)
		} else {
			last = f
		}
		size += len(f)
	}
	var n int
	var buf = b.slice.Make(size)
	for _, frag := range full {
		n += copy(buf[n:], frag)
	}
	copy(buf[n:], last)
	return buf, nil
}
func (b *Reader) ReadSlice(delim byte) ([]byte, error) {
	if b.err != nil {
		return nil, b.err
	}
	for {
		var index = bytes.IndexByte(b.buf[b.rpos:b.wpos], delim)
		if index >= 0 {
			limit := b.rpos + index + 1
			slice := b.buf[b.rpos:limit]
			b.rpos = limit
			return slice, nil
		}
		if b.buffered() == len(b.buf) {
			b.rpos = b.wpos
			return b.buf, bufio.ErrBufferFull
		}
		if b.fill() != nil {
			return nil, b.err
		}
	}
}

func (b *Reader) ReadFull(n int) ([]byte, error) {
	
	if b.err != nil || n == 0 {
		return nil, b.err
	}
	var buf = b.slice.Make(n)
	if _, err := io.ReadFull(bytes.NewReader(b.buf[b.rpos:]), buf); err != nil {
		return nil, err
	}
	b.rpos += n
	return buf, nil
}


func (d *sliceAlloc) Make(n int) (ss []byte) {
	switch {
	case n == 0:
		return []byte{}
	case n >= 512:
		return make([]byte, n)
	default:
		if len(d.buf) < n {
			d.buf = make([]byte, 8192)
		}
		ss, d.buf = d.buf[:n:n], d.buf[n:]
		return ss
	}
}

func (b *Reader) PeekByte() (byte, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.buffered() == 0 {
		if b.fill() != nil {
			return 0, b.err
		}
	}
	c := b.buf[b.rpos]
	return c, nil
}