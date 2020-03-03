package proto

import (
	"bytes"
	"gReKVdisDB/utils"
	"gReKVdisDB/utils/gbufio"
	"io"
	"strconv"
)

type Encoder struct {
	bw *gbufio.Writer
	Err error
}

func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderBuffer(gbufio.NewWriterSize(w, 8192))
}

func NewEncoderSize(w io.Writer, size int) *Encoder {
	return NewEncoderBuffer(gbufio.NewWriterSize(w, size))
}


func NewEncoderBuffer(bw *gbufio.Writer) *Encoder {
	return &Encoder{bw: bw}
}

func (e *Encoder) Encode(r *Resp, flush bool) error {
	if e.Err != nil {
		return utils.ErrorsTrace(e.Err)
	}
	if err := e.encodeResp(r); err != nil {
		e.Err = err
	} else if flush {
		e.Err = utils.ErrorsTrace(e.bw.Flush())
	}
	return e.Err
}

func EncodeCmd(cmd string) ([]byte, error) {
	return EncodeBytes([]byte(cmd))
}

func EncodeBytes(b []byte) ([]byte, error) {
	r := bytes.Split(b, []byte(" "))
	if r == nil {
		return nil, utils.ErrorsTrace(utils.ErrorNew("empty split"))
	}
	resp := NewArray(nil)
	for _, v := range r {
		if len(v) > 0 {
			resp.Array = append(resp.Array, NewBulkBytes(v))
		}
	}
	return EncodeToBytes(resp)
}

func (e *Encoder) EncodeMultiBulk(multi []*Resp, flush bool) error {
	if e.Err != nil {
		return utils.ErrorsTrace(e.Err)
	}
	if err := e.encodeMultiBulk(multi); err != nil {
		e.Err = err
	} else if flush {
		e.Err = utils.ErrorsTrace(e.Err)
	}
	return e.Err
}

func (e *Encoder) Flush() error {
	if e.Err != nil {
		return utils.ErrorsTrace(utils.ErrorNew("Flush error"))
	}
	if err := e.bw.Flush(); err != nil {
		e.Err = utils.ErrorsTrace(utils.ErrorNew("bw.Flush error"))
	}
	return e.Err
}

func Encode(w io.Writer, r *Resp) error {
	return NewEncoder(w).Encode(r, true)
}


func EncodeToBytes(r *Resp) ([]byte, error) {
	var b = &bytes.Buffer{}
	if err := Encode(b, r); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *Encoder) encodeResp(r *Resp) error {
	if err := e.bw.WriteByte(byte(r.Type)); err != nil {
		return utils.ErrorsTrace(err)
	}
	switch r.Type {
	case TypeString, TypeError, TypeInt:
		return e.encodeTextBytes(r.Value)
	case TypeBulkBytes:
		return e.encodeBulkBytes(r.Value)
	case TypeArray:
		return e.encodeArray(r.Array)
	default:
		return utils.ErrorsTrace(e.Err)
	}
}

func (e *Encoder) encodeMultiBulk(multi []*Resp) error {
	if err := e.bw.WriteByte(byte(TypeArray)); err != nil {
		return utils.ErrorsTrace(err)
	}
	return e.encodeArray(multi)
}

func (e *Encoder) encodeTextBytes(b []byte) error {
	if _, err := e.bw.Write(b); err != nil {
		return utils.ErrorsTrace(err)
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return utils.ErrorsTrace(err)
	}
	return nil
}

func (e *Encoder) encodeTextString(s string) error {
	if _, err := e.bw.WriteString(s); err != nil {
		return utils.ErrorsTrace(err)
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return utils.ErrorsTrace(err)
	}
	return nil
}

func (e *Encoder) encodeInt(v int64) error {
	return e.encodeTextString(strconv.FormatInt(v, 10))
}

func (e *Encoder) encodeBulkBytes(b []byte) error {
	if b == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(b))); err != nil {
			return err
		}
		return e.encodeTextBytes(b)
	}
}

func (e *Encoder) encodeArray(array []*Resp) error {
	if array == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(array))); err != nil {
			return err
		}
		for _, r := range array {
			if err := e.encodeResp(r); err != nil {
				return err
			}
		}
		return nil
	}
}