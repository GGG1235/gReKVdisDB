package core

import (
	"gReKVdisDB/date_structure/skiplist"
	"gReKVdisDB/utils"
	"math/rand"
)


const ADD_NONE = 0
const ADD_INCR = (1 << 0) 
const ADD_NX = (1 << 1)   
const ADD_XX = (1 << 2)


const ADD_NOP = (1 << 3)     
const ADD_NAN = (1 << 4)     
const ADD_ADDED = (1 << 5)   
const ADD_UPDATED = (1 << 6) 

const SKIPLIST_MAXLEVEL = 32
const SKIPLIST_P = 0.25 

type Set struct {
	dict *Dict
	Sl  *skiplist.SkipList
}

type RangeSpec struct {
	Min   float64
	Max   float64
	MinEx int
	MaxEx int
}

func addCommand(c *Client) {
	addGenericCommand(c, ADD_NONE)
}

func IncrbyCommand(c *Client) {
	addGenericCommand(c, ADD_INCR)
}

func addGenericCommand(c *Client, flags int) {
	key := c.Argv[1]
	scoreIdx := 2
	elements := c.Argc - scoreIdx
	elements /= 2
	scores := make([]float64, elements)

	for j := 0; j < elements; j++ {
		if value, ok := c.Argv[2+j*2].Ptr.(uint64); ok {
			scores[j] = float64(value)
		}
	}


	obj := LookupKey(c.Db, key)
	if obj == nil {

		obj = createSetObject()

		c.Db.Dict[key.Ptr.(string)] = obj
	}

	for j := 0; j < elements; j++ {
		var newScore float64
		Score := scores[j]
		retFlags := flags
		if ele, ok := c.Argv[scoreIdx+1+j*2].Ptr.(string); ok {
			SetAdd(obj, Score, ele, &retFlags, &newScore)
		}

	}

}


func createSetObject() *utils.GKVDBObject {
	val := new(Set)
	val.dict = new(Dict)
	dict := make(map[string]*utils.GKVDBObject)
	*val.dict = dict

	val.Sl = skiplist.CreateSkipList()
	o := utils.CreateObject(utils.OBJ_SET, val)
	return o
}

func SetAdd(Obj *utils.GKVDBObject, Score float64, Ele string, flags *int, newScore *float64) bool {
	incr := (*flags & ADD_INCR) != 0
	nx := (*flags & ADD_NX) != 0
	xx := (*flags & ADD_XX) != 0
	*flags = 0
	var curscore float64

	if Obj.ObjectType == utils.OBJ_SET {

		s := Obj.Ptr.(*Set)

		dict := s.dict
		de := dictFind(dict, Ele)
		if de != nil {
			if nx {

			}
			if incr {

			}

			if coreTemp, ok := de.Ptr.(float64); ok {
				curscore = coreTemp
			} else {

			}


			if curscore != Score {

			}

		} else if !xx {

			Insert(s.Sl, Score, Ele)

			(*(s.dict))[Ele] = utils.CreateObject(utils.ObjectTypeString, Score)
			*flags |= ADD_ADDED
			return true
		}
	} else {

	}
	return false
}

func dictFind(d *Dict, key string) *utils.GKVDBObject {
	if (*d)[key] != nil {
		return (*d)[key]
	}
	return nil
}

func Insert(Sl *skiplist.SkipList, Score float64, Ele string) *skiplist.Node {
	update := make([]*skiplist.Node, SKIPLIST_MAXLEVEL)
	rank := make([]uint, SKIPLIST_MAXLEVEL)
	x := Sl.Header

	for i := Sl.Level - 1; i >= 0; i-- {
		if i == Sl.Level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.Level[i].Forward != nil && (x.Level[i].Forward.Score < Score ||
			(x.Level[i].Forward.Score == Score && (x.Level[i].Forward.Ele < Ele))) {
			rank[i] += x.Level[i].Span
			x = x.Level[i].Forward
		}
		update[i] = x
	}

	Level := RandomLevel()
	if Level > Sl.Level {
		for i := Sl.Level; i < Level; i++ {
			rank[i] = 0
			update[i] = Sl.Header
			update[i].Level[i].Span = Sl.Length
		}
		Sl.Level = Level
	}

	x = skiplist.CreateSkipListNode(Level, Score, Ele)
	for i := 0; i < Level; i++ {
		x.Level[i].Forward = update[i].Level[i].Forward
		update[i].Level[i].Forward = x
		x.Level[i].Span = update[i].Level[i].Span - (rank[0] - rank[i])
		update[i].Level[i].Span = (rank[0] - rank[i]) + 1
	}

	for i := Level; i < Sl.Level; i++ {
		update[i].Level[i].Span++
	}

	if update[0] == Sl.Header {
		x.Backward = nil
	} else {
		x.Backward = update[0]
	}

	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x
	} else {
		Sl.Tail = x
	}
	Sl.Length++
	return x
}


func RandomLevel() int {
	Level := 1
	for rand.Float64()*65535 < SKIPLIST_P*65535 {
		Level++
	}

	if Level < SKIPLIST_MAXLEVEL {
		return Level
	}
	return SKIPLIST_MAXLEVEL
}

func SkipListFirstInRange(Sl *skiplist.SkipList, Range *RangeSpec) *skiplist.Node {
	if !IsInRange(Sl, Range) {
		return nil
	}
	x := Sl.Header
	for i := Sl.Level - 1; i >= 0; i-- {
		for x.Level[i].Forward != nil && !SkipListValueGteMin(x.Level[i].Forward.Score, Range) {
			x = x.Level[i].Forward
		}
	}

	x = x.Level[0].Forward
	if x == nil {
		return nil
	}

	if !SkipListValueLteMax(x.Score, Range) {
		return nil
	}
	return x
}

func SkipListValueGteMin(value float64, spec *RangeSpec) bool {
	if spec.MinEx != 0 {
		return value > spec.Min
	}
	return value >= spec.Min
}

func SkipListValueLteMax(value float64, spec *RangeSpec) bool {
	if spec.MaxEx != 0 {
		return value < spec.Max
	}
	return value <= spec.Max
}

func IsInRange(Sl *skiplist.SkipList, Range *RangeSpec) bool {

	if Range.Min > Range.Max ||
		(Range.Min == Range.Max && (Range.MinEx != 0 || Range.MaxEx != 0)) {
		return false
	}
	x := Sl.Tail
	if x == nil || !SkipListValueGteMin(x.Score, Range) {
		return false
	}
	x = Sl.Header.Level[0].Forward
	if x == nil || !SkipListValueLteMax(x.Score, Range) {
		return false
	}
	return true
}


func Delete(Sl *skiplist.SkipList, Score float64, Ele string, node **skiplist.Node) bool {
	update := make([]*skiplist.Node, SKIPLIST_MAXLEVEL)
	rank := make([]uint, SKIPLIST_MAXLEVEL)
	x := Sl.Header
	for i := Sl.Level - 1; i >= 0; i-- {
		if i == Sl.Level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.Level[i].Forward != nil && (x.Level[i].Forward.Score < Score ||
			(x.Level[i].Forward.Score == Score && (x.Level[i].Forward.Ele < Ele))) {
			rank[i] += x.Level[i].Span
			x = x.Level[i].Forward
		}
		update[i] = x
	}
	x = x.Level[0].Forward
	if x != nil && Score == x.Score && Ele == x.Ele {
		DeleteNode(Sl, x, update)
		return true
	}
	return false
}

func DeleteNode(Sl *skiplist.SkipList, x *skiplist.Node, update []*skiplist.Node) {
	for i := 0; i < Sl.Level; i++ {
		if update[i].Level[i].Forward == x {
			update[i].Level[i].Span += x.Level[i].Span - 1
			update[i].Level[i].Forward = x.Level[i].Forward
		} else {
			update[i].Level[i].Span -= 1
		}
	}

	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x.Backward
	} else {
		Sl.Tail = x.Backward
	}

	for Sl.Level > 1 && Sl.Header.Level[Sl.Level-1].Forward == nil {
		Sl.Level--
	}
	Sl.Length--
}

func SetScore(obj *utils.GKVDBObject, member string, Score *float64) int {
	if obj == nil || member == "" {
		return utils.C_ERR
	}

	if obj.ObjectType == utils.OBJ_SET {
		s := obj.Ptr.(*Set)
		dict := s.dict
		de := dictFind(dict, member)

		if de == nil {
			return utils.C_ERR
		}
		value := de.Ptr.(float64)
		*Score = value
	} else {
		panic("Unknown sorted set encoding")
	}
	return utils.C_OK
}
