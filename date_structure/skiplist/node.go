package skiplist

type Node struct {
	Ele      string
	Score    float64
	Backward *Node
	Level    []Level
}

type Level struct {
	Forward *Node
	Span    uint
}

func CreateSkipListNode(level int, Score float64, Ele string) *Node {
	zn := new(Node)
	zl := make([]Level, level)
	zn.Level = zl
	zn.Score = Score
	zn.Ele = Ele
	return zn
}