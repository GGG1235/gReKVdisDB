package core

import "gReKVdisDB/utils"

type GkvdisDb struct {
	Dict    Dict
	Expires Dict
	ID      int32
}

func LookupKey(db *GkvdisDb, key *utils.GKVDBObject) (ret *utils.GKVDBObject) {
	if o, ok := db.Dict[key.Ptr.(string)]; ok {
		return o
	}
	return nil
}