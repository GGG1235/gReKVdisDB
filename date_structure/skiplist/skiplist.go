package skiplist

const SKIPLIST_MAXLEVEL = 32
const SKIPLIST_P = 0.25

type SkipList struct {
	Header *Node
	Tail   *Node
	Length uint
	Level  int
}

func CreateSkipList() *SkipList {
	zsl := new(SkipList)
	zsl.Level = 1
	zsl.Length = 0
	zsl.Header = CreateSkipListNode(SKIPLIST_MAXLEVEL, 0, "")
	for j := 0; j < SKIPLIST_MAXLEVEL; j++ {
		zsl.Header.Level[j].Forward = nil
		zsl.Header.Level[j].Span = 0
	}
	zsl.Header.Backward = nil
	zsl.Tail = nil
	return zsl
}