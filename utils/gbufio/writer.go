package gbufio

import "io"

type Writer struct {
	err error
	buf []byte

	wr   io.Writer
	wpos int
}

func NewWriter(wr io.Writer) *Writer {
	return NewWriterSize(wr, 1024)
}

func NewWriterSize(wr io.Writer, size int) *Writer {
	if size <= 0 {
		size = 1024
	}
	return &Writer{wr: wr, buf: make([]byte, size)}
}

func (w *Writer) Flush() error {
	return w.flush()
}

func (w *Writer) flush() error {
	if w.err != nil {
		return w.err
	}
	if w.wpos == 0 {
		return nil
	}
	n, err := w.wr.Write(w.buf[:w.wpos])
	if err != nil {
		w.err = err
	} else if n < w.wpos {
		w.err = io.ErrShortWrite
	} else {
		w.wpos = 0
	}
	return w.err
}

func (w *Writer) available() int {
	return len(w.buf) - w.wpos
}
func (w *Writer) Write(p []byte) (nn int, err error) {
	for w.err == nil && len(p) > w.available() {
		var n int
		if w.wpos == 0 {
			n, w.err = w.wr.Write(p)
		} else {
			n = copy(w.buf[w.wpos:], p)
			w.wpos += n
			w.flush()
		}
		nn, p = nn+n, p[n:]
	}
	if w.err != nil || len(p) == 0 {
		return nn, w.err
	}
	n := copy(w.buf[w.wpos:], p)
	w.wpos += n
	return nn + n, nil
}


func (w *Writer) WriteByte(c byte) error {
	if w.err != nil {
		return w.err
	}
	if w.available() == 0 && w.flush() != nil {
		return w.err
	}
	w.buf[w.wpos] = c
	w.wpos++
	return nil
}

func (w *Writer) WriteString(s string) (nn int, err error) {
	for w.err == nil && len(s) > w.available() {
		n := copy(w.buf[w.wpos:], s)
		w.wpos += n
		w.flush()
		nn, s = nn+n, s[n:]
	}
	if w.err != nil || len(s) == 0 {
		return nn, w.err
	}
	n := copy(w.buf[w.wpos:], s)
	w.wpos += n
	return nn + n, nil
}