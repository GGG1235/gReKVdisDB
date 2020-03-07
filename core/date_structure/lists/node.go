package lists

type listNode struct {
	Prev  *listNode
	Next  *listNode
	Value interface{}
}

func (n listNode) listPrevNode() *listNode {
	return n.Prev
}

func (n listNode) listNextNode() *listNode {
	return n.Next
}

func (n listNode) listNodeValue() interface{} {
	return n.Value
}