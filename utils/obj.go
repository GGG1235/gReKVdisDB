package utils

type GKVDBObject struct {
	ObjectType int
	Ptr interface{}
}

const C_ERR = -1
const C_OK = 0

const ObjectTypeString = 0
const OBJ_LIST = 1
const OBJ_SET = 2
const OBJ_ZSET = 3
const OBJ_HASH = 4

const OBJ_ENCODING_RAW = 0        
const OBJ_ENCODING_INT = 1        
const OBJ_ENCODING_HT = 2         
const OBJ_ENCODING_ZIPMAP = 3     
const OBJ_ENCODING_LINKEDLIST = 4 
const OBJ_ENCODING_ZIPLIST = 5    
const OBJ_ENCODING_INTSET = 6     
const OBJ_ENCODING_SKIPLIST = 7   
const OBJ_ENCODING_EMBSTR = 8     
const OBJ_ENCODING_QUICKLIST = 9  

func CreateObject(t int, ptr interface{}) (o *GKVDBObject) {
	o = new(GKVDBObject)
	o.ObjectType = t
	o.Ptr = ptr
	return
}
