package lists

type List struct {
	Head *listNode
	Tail *listNode
	Len  int
}

func (l List) ListLength() int {
	return l.Len
}

func (l List) ListFirst() *listNode {
	return l.Head
}

func (l List) ListLast() *listNode {
	return l.Tail
}

func ListCreate() *List {
	list := new(List)
	list.Head = nil
	list.Tail = nil
	list.Len = 0
	return list
}

func (l *List) ListAddNodeHead(Value interface{}) *List {
	node := new(listNode)

	node.Value = Value
	if l.Len == 0 {
		l.Head = node
		l.Tail = node
		node.Prev = nil
		node.Next = nil
	} else {
		node.Prev = nil
		node.Next = l.Head
		l.Head.Prev = node
		l.Head = node
	}
	l.Len++
	return l
}

func (l *List) ListAddNodeTail(Value interface{}) *List {
	node := new(listNode)

	node.Value = Value
	if l.Len == 0 {
		l.Head = node
		l.Tail = node
		node.Prev = nil
		node.Next = nil
	} else {
		node.Prev = l.Tail
		node.Next = nil
		l.Tail.Next = node
		l.Tail = node
	}
	l.Len++
	return l
}

func (l *List) ListInsertNode(oldNode *listNode, Value interface{}, after int) *List {
	node := new(listNode)
	node.Value = Value
	if after > 0 {

	}
	l.Len++
	return l
}
