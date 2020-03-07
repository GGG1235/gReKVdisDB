package core

import "gReKVdisDB/utils"

type GkvdisDb struct {
	Dict    dict
	Expires dict
	ID      int32
}

func lookupKey(db *GkvdisDb, key *utils.GKVDBObject) (ret *utils.GKVDBObject) {
	if o, ok := db.Dict[key.Ptr.(string)]; ok {
		return o
	}
	return nil
}