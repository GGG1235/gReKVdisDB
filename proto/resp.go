package proto

type Resp struct {
	Type byte

	Value []byte
	Array []*Resp
}

func NewString(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeString
	r.Value = value
	return r
}

func NewError(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeError
	r.Value = value
	return r
}

func NewInt(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeInt
	r.Value = value
	return r
}

func NewBulkBytes(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeBulkBytes
	r.Value = value
	return r
}


func NewArray(array []*Resp) *Resp {
	r := &Resp{}
	r.Type = TypeArray
	r.Array = array
	return r
}
