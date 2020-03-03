package proto

import (
	"bytes"
	"gReKVdisDB/utils"
	"gReKVdisDB/utils/gbufio"
	"io"
)

type Decoder struct {
	br *gbufio.Reader
	Err error
}

func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderBuffer(gbufio.NewReaderSize(r, 8192))
}

func NewDecoderSize(r io.Reader, size int) *Decoder {
	return NewDecoderBuffer(gbufio.NewReaderSize(r, size))
}

func NewDecoderBuffer(br *gbufio.Reader) *Decoder {
	return &Decoder{br: br}
}

func (d *Decoder) Decode() (*Resp, error) {
	if d.Err != nil {
		return nil, utils.ErrorsTrace(utils.ErrorNew("Decode err"))
	}
	r, err := d.decodeResp()
	if err != nil {
		d.Err = err
	}
	return r, d.Err
}

func (d *Decoder) DecodeMultiBulk() ([]*Resp, error) {
	if d.Err != nil {
		return nil, utils.ErrorsTrace(utils.ErrorNew("DecodeMultibulk error"))
	}
	m, err := d.decodeMultiBulk()
	if err != nil {
		d.Err = err
	}
	return m, err
}

func Decode(r io.Reader) (*Resp, error) {
	return NewDecoder(r).Decode()
}

func DecodeFromBytes(p []byte) (*Resp, error) {
	return NewDecoder(bytes.NewReader(p)).Decode()
}

func DecodeMultiBulkFromBytes(p []byte) ([]*Resp, error) {
	return NewDecoder(bytes.NewReader(p)).DecodeMultiBulk()
}

func (d *Decoder) decodeResp() (*Resp, error) {
	b, err := d.br.ReadByte()
	if err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	r := &Resp{}
	r.Type = byte(b)
	switch r.Type {
	default:
		return nil, utils.ErrorsTrace(err)
	case TypeString, TypeError, TypeInt:
		r.Value, err = d.decodeTextBytes()
	case TypeBulkBytes:
		r.Value, err = d.decodeBulkBytes()
	case TypeArray:
		r.Array, err = d.decodeArray()
	}
	return r, err
}

func (d *Decoder) decodeTextBytes() ([]byte, error) {
	b, err := d.br.ReadBytes('\n')
	if err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	if n := len(b) - 2; n < 0 || b[n] != '\r' {
		return nil, utils.ErrorsTrace(err)
	} else {
		return b[:n], nil
	}
}

func (d *Decoder) decodeInt() (int64, error) {
	b, err := d.br.ReadSlice('\n')
	if err != nil {
		return 0, utils.ErrorsTrace(err)
	}
	if n := len(b) - 2; n < 0 || b[n] != '\r' {
		return 0, utils.ErrorsTrace(err)
	} else {
		return Btoi64(b[:n])
	}
}

func (d *Decoder) decodeBulkBytes() ([]byte, error) {
	n, err := d.decodeInt()
	if err != nil {
		return nil, err
	}
	switch {
	case n < -1:
		return nil, utils.ErrorsTrace(err)
	case n > MaxBulkBytesLen:
		return nil, utils.ErrorsTrace(err)
	case n == -1:
		return nil, nil
	}
	b, err := d.br.ReadFull(int(n) + 2)
	if err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	if b[n] != '\r' || b[n+1] != '\n' {
		return nil, utils.ErrorsTrace(err)
	}
	return b[:n], nil
}

func (d *Decoder) decodeArray() ([]*Resp, error) {
	n, err := d.decodeInt()
	if err != nil {
		return nil, err
	}
	switch {
	case n < -1:
		return nil, utils.ErrorsTrace(err)
	case n > MaxArrayLen:
		return nil, utils.ErrorsTrace(err)
	case n == -1:
		return nil, nil
	}
	array := make([]*Resp, n)
	for i := range array {
		r, err := d.decodeResp()
		if err != nil {
			return nil, err
		}
		array[i] = r
	}
	return array, nil
}

func (d *Decoder) decodeSingleLineMultiBulk() ([]*Resp, error) {
	b, err := d.decodeTextBytes()
	if err != nil {
		return nil, err
	}
	multi := make([]*Resp, 0, 8)
	for l, r := 0, 0; r <= len(b); r++ {
		if r == len(b) || b[r] == ' ' {
			if l < r {
				multi = append(multi, NewBulkBytes(b[l:r]))
			}
			l = r + 1
		}
	}
	if len(multi) == 0 {
		return nil, utils.ErrorsTrace(err)
	}
	return multi, nil
}

func (d *Decoder) decodeMultiBulk() ([]*Resp, error) {
	b, err := d.br.PeekByte()
	if err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	if RespType(b) != TypeArray {
		return d.decodeSingleLineMultiBulk()
	}
	if _, err := d.br.ReadByte(); err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	n, err := d.decodeInt()

	if err != nil {
		return nil, utils.ErrorsTrace(err)
	}
	switch {
	case n <= 0:
		return nil, utils.ErrorsTrace(utils.ErrBadArrayLen)
	case n > MaxArrayLen:
		return nil, utils.ErrorsTrace(utils.ErrBadArrayLenTooLong)
	}
	multi := make([]*Resp, n)
	for i := range multi {
		r, err := d.decodeResp()
		if err != nil {
			return nil, err
		}
		if r.Type != TypeBulkBytes {
			return nil, utils.ErrorsTrace(utils.ErrBadMultiBulkContent)
		}
		multi[i] = r
	}
	return multi, nil
}