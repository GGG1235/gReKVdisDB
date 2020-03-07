package core

import (
	"fmt"
	"gReKVdisDB/utils"
	"os"
	"strconv"
	"strings"
)

const RADIUS_COORDS = (1 << 0)
const RADIUS_MEMBER = (1 << 1)
const RADIUS_NOSTORE = (1 << 2)

const SORT_NONE = 0
const SORT_ASC = 1
const SORT_DESC = 2


func AddCommand(c *Client, s *Server) {
	
	if (c.Argc-2)%3 != 0 {
		addReplyError(c, "syntax error. Try GEOADD key [x1] [y1] [name1] "+
			"[x2] [y2] [name2] ... ")
	}

	elements := (c.Argc - 2) / 3 
	argc := 2 + elements*2 
	argv := make([]*utils.GKVDBObject, argc)
	argv[0] = utils.CreateObject(utils.ObjectTypeString, "zadd")
	argv[1] = c.Argv[1]

	for i := 0; i < elements; i++ {
		var xy [2]float64
		var hash HashBits
		
		if lngObj, ok1 := c.Argv[i*3+2].Ptr.(string); ok1 {
			if latObj, ok2 := c.Argv[i*3+3].Ptr.(string); ok2 {
				var ok error
				xy[0], ok = strconv.ParseFloat(lngObj, 64)
				xy[1], ok = strconv.ParseFloat(latObj, 64)
				if ok != nil {
					addReplyError(c, "lng lat type error")
					os.Exit(0)
				}
			}
		}
		hashEncodeWGS84(xy[0], xy[1], GEO_STEP_MAX, &hash)
		bits := hashAlign52Bits(hash)
		score := utils.CreateObject(utils.ObjectTypeString, bits)

		val := c.Argv[2+i*3+2]
		argv[2+i*2] = score 
		argv[3+i*2] = val
	}
	c.Argc = argc
	c.Argv = argv
	zaddCommand(c)

	addReplyStatus(c, "OK")
}


func HashCommand(c *Client, s *Server) {
	Alphabet := "0123456789bcdefghjkmnpqrstuvwxyz"
	zobj := lookupKey(c.Db, c.Argv[1])
	if zobj != nil && zobj.ObjectType != utils.OBJ_ZSET {
		return
	}
	buf := ""
	for j := 2; j < c.Argc; j++ {
		var score float64
		if zobj == nil || zsetScore(zobj, c.Argv[j].Ptr.(string), &score) == utils.C_ERR {
			addReplyError(c, "score get error ")
			return
		}
		var xy [2]float64
		if !decodehash(score, &xy) {
			addReplyError(c, "hash get error")
			continue
		}
		r := [2]HashRange{}
		var hash HashBits
		r[0].min = -180
		r[0].max = 180
		r[1].min = -90
		r[1].max = 90
		hashEncode(&r[0], &r[1], xy[0], xy[1], 26, &hash)

		temp := ""
		for i := 0; i < 11; i++ {
			count := 52 - (i+1)*5
			idx := (hash.bits >> (uint(count))) & 0x1f
			temp += string(Alphabet[idx])
		}
		buf += temp
		buf += ";"
	}
	addReplyStatus(c, buf)
}


func PosCommand(c *Client, s *Server) {
	zobj := lookupKey(c.Db, c.Argv[1])
	if zobj != nil && zobj.ObjectType != utils.OBJ_ZSET {
		return
	}
	buf := "lng:"

	for j := 2; j < c.Argc; j++ {
		var score float64
		if zobj == nil || zsetScore(zobj, c.Argv[j].Ptr.(string), &score) == utils.C_ERR {
			addReplyError(c, "score get error ")
			return
		}
		var xy [2]float64
		if !decodehash(score, &xy) {
			addReplyError(c, "hash get error")
			continue
		}

		buf += fmt.Sprint(xy[0])
		buf += ",lat:"
		buf += fmt.Sprint(xy[1])
		buf += ";"
	}
	addReplyStatus(c, buf)
}


func DistCommand(c *Client, s *Server) {
	if c.Argc >= 5 {
		addReplyError(c, "params error")
		return
	}
	zobj := lookupKey(c.Db, c.Argv[1])
	if zobj != nil && zobj.ObjectType != utils.OBJ_ZSET {
		return
	}

	var score1, score2 float64
	var xyxy1, xyxy2 [2]float64
	if zsetScore(zobj, c.Argv[2].Ptr.(string), &score1) == utils.C_ERR ||
		zsetScore(zobj, c.Argv[3].Ptr.(string), &score2) == utils.C_ERR {
		addReplyError(c, "score get error ")
		return
	}

	if !decodehash(score1, &xyxy1) || !decodehash(score2, &xyxy2) {
		addReplyError(c, "hash get error")
		return
	}

	buf := hashGetDistance(xyxy1[0], xyxy1[1], xyxy2[0], xyxy2[1])
	addReplyStatus(c, fmt.Sprint(buf))
}

func RadiusCommand(c *Client, s *Server) {
	radiusGeneric(c, RADIUS_COORDS)
}

func RadiusByMemberCommand(c *Client, s *Server) {
	radiusGeneric(c, RADIUS_MEMBER)
}


func radiusGeneric(c *Client, flags uint) {
	var storekey *utils.GKVDBObject
	storedist := 0

	
	zobj := lookupKey(c.Db, c.Argv[1])
	if zobj != nil && zobj.ObjectType != utils.OBJ_ZSET {
		return
	}

	var xy [2]float64
	var base_args int
	if flags&RADIUS_COORDS > 0 {
		base_args = 6
		arg2, ok1 := c.Argv[2].Ptr.(string)
		arg3, ok2 := c.Argv[3].Ptr.(string)
		if !ok1 || !ok2 {
			addReplyError(c, "get lng lat error")
			return
		}

		var err error
		xy[0], err = strconv.ParseFloat(arg2, 64)
		xy[1], err = strconv.ParseFloat(arg3, 64)
		if err != nil {
			addReplyError(c, "get lng lat float error")
			return
		}
	} else if flags&RADIUS_MEMBER > 0 {
		
		base_args = 7
	} else {
		addReplyError(c, "Unknown radius search type")
		return
	}

	
	conversion := extractUnitOrReply(c, *c.Argv[base_args-1])
	radius_meters, err := strconv.ParseFloat(c.Argv[base_args-2].Ptr.(string), 64)
	if err != nil {
		addReplyError(c, "radius_meters error")
		return
	}
	radius_meters = radius_meters * conversion

	
	withdist := 0
	withhash := 0
	withcoords := 0
	sort := SORT_NONE
	var count int64 = 0
	if c.Argc > base_args {
		remaining := c.Argc - base_args
		for i := 0; i < remaining; i++ {
			arg := c.Argv[base_args+i].Ptr.(string)
			if strings.EqualFold(arg, "withdist") {
				withdist = 1
			} else if strings.EqualFold(arg, "withhash") {
				withhash = 1
			} else if strings.EqualFold(arg, "withcoord") {
				withcoords = 1
			} else if strings.EqualFold(arg, "asc") {
				sort = SORT_ASC
			} else if strings.EqualFold(arg, "desc") {
				sort = SORT_DESC
			} else if strings.EqualFold(arg, "count") && (i+1) < remaining {

				if count < 0 {
					addReplyError(c, "COUNT must be > 0")
					return
				}
				i++
			} else if strings.EqualFold(arg, "store") && (i+1) < remaining && (flags&RADIUS_NOSTORE == 0) {
				storekey = c.Argv[base_args+i+1]
				storedist = 0
				i++
			} else if strings.EqualFold(arg, "storedist") && (i+1) < remaining && (flags&RADIUS_NOSTORE == 0) {
				storekey = c.Argv[base_args+i+1]
				storedist = 1
				i++
			} else {
				addReplyError(c, "params error")
				return
			}
		}
	}

	if storekey != nil && (withdist > 0 || withhash > 0 || withcoords > 0) {
		addReplyError(c,
			"STORE option in GEORADIUS is not compatible with "+
				"WITHDIST, WITHHASH and WITHCOORDS options")
		return
	}

	
	if count != 0 && sort == SORT_NONE {
		sort = SORT_ASC
	}

	
	radius := hashGetAreasByRadiusWGS84(xy[0], xy[1], radius_meters)

	/* Search the zset for all matching points */
	ga := ArrayCreate() 
	membersOfAllNeighbors(zobj, radius, xy[0], xy[1], radius_meters, ga)

	if ga.used == 0 && storekey == nil {
		addReplyError(c, "emptymultibulk")
		return
	}

	result_length := ga.used
	var returned_items int
	if count == 0 || int64(result_length) < count {
		returned_items = int(result_length)
	} else {
		returned_items = int(count)
	}
	option_length := 0

	if sort == SORT_ASC {

	} else if sort == SORT_DESC {

	}

	if storekey == nil {
		if withdist > 0 {
			option_length++
		}
		if withcoords > 0 {
			option_length++
		}
		if withhash > 0 {
			option_length++
		}

		/* Finally send results back to the caller */
		for i := 0; i < returned_items; i++ {
			gp := ga.array[i]
			gp.dist /= conversion
			fmt.Println(gp)
			addReplyStatus(c, gp.member)
		}

	} else {
		fmt.Println(storedist)
	}

}

func ArrayCreate() *Array {
	ga := new(Array)
	ga.array = make([]*Point, 0)
	ga.buckets = 0
	ga.used = 0
	return ga
}


func extractUnitOrReply(c *Client, uint utils.GKVDBObject) float64 {
	u := uint.Ptr.(string)

	if strings.Compare(u, "m") == 0 {
		return 1
	} else if strings.Compare(u, "km") == 0 {
		return 1000
	} else if strings.Compare(u, "ft") == 0 {
		return 0.3048
	} else if strings.Compare(u, "mi") == 0 {
		return 1609.34
	} else {
		addReplyError(c, "unsupported unit provided. please use m, km, ft, mi")
		return -1
	}
}

func membersOfAllNeighbors(zobj *utils.GKVDBObject, n HashRadius, lon float64, lat float64, radius float64, ga *Array) int {
	neighbors := [9]HashBits{}
	var count, last_processed int
	debugmsg := 0

	neighbors[0] = n.hash
	neighbors[1] = n.neighbors.north
	neighbors[2] = n.neighbors.south
	neighbors[3] = n.neighbors.east
	neighbors[4] = n.neighbors.west
	neighbors[5] = n.neighbors.north_east
	neighbors[6] = n.neighbors.north_west
	neighbors[7] = n.neighbors.south_east
	neighbors[8] = n.neighbors.south_west

	for i := 0; i < len(neighbors); i++ {
		if hashIsZero(neighbors[i]) {
			continue
		}

		/* Debugging info. */
		if debugmsg > 0 {
			var long_range, lat_range HashRange
			hashGetCoordRange(&long_range, &lat_range)
			myarea := new(HashArea)
			hashDecode(long_range, lat_range, neighbors[i], myarea)

			/* Dump center square. */
			fmt.Println("neighbors[%d]:\n", i)
			fmt.Println("area.longitude.min: %f\n", myarea.longitude.min)
			fmt.Println("area.longitude.max: %f\n", myarea.longitude.max)
			fmt.Println("area.latitude.min: %f\n", myarea.latitude.min)
			fmt.Println("area.latitude.max: %f\n", myarea.latitude.max)
		}

		/* When a huge Radius (in the 5000 km range or more) is used,
		 * adjacent neighbors can be the same, leading to duplicated
		 * elements. Skip every range which is the same as the one
		 * processed previously. */
		if last_processed > 0 &&
			neighbors[i].bits == neighbors[last_processed].bits &&
			neighbors[i].step == neighbors[last_processed].step {
			if debugmsg > 0 {
				fmt.Println("Skipping processing of %d, same as previous\n", i)
			}
			continue
		}
		count += membersOfHashBox(zobj, neighbors[i], ga, lon, lat, radius)
		last_processed = i
	}
	return count
}

func membersOfHashBox(zobj *utils.GKVDBObject, hash HashBits, ga *Array, lon float64, lat float64, radius float64) int {
	var min, max HashFix52Bits

	scoresOfHashBox(hash, &min, &max)
	return GetPointsInRange(zobj, float64(min), float64(max), lon, lat, radius, ga)
}

func scoresOfHashBox(hash HashBits, min *HashFix52Bits, max *HashFix52Bits) {
	*min = hashAlign52Bits(hash)
	hash.bits++
	*max = hashAlign52Bits(hash)
}

func GetPointsInRange(zobj *utils.GKVDBObject, min float64, max float64, lon float64, lat float64, radius float64, ga *Array) int {
	zrange := zRangeSpec{min: min, max: max, minEx: 0, maxEx: 1}
	var origincount uint = ga.used
	
	if zobj.ObjectType == utils.OBJ_ZSET {
		zs := zobj.Ptr.(*zSet) 
		zsl := zs.zsl
		var ln *zSkipListNode

		ln = zslFirstInRange(zsl, &zrange)
		if ln == nil {
			return 0
		}

		for ln != nil {
			ele := ln.ele
			if !zslValueLteMax(ln.score, &zrange) {
				break
			}
			AppendIfWithinRadius(ga, lon, lat, radius, ln.score, ele)
			ln = ln.level[0].forward
		}
	} else {
		
	}
	return int(ga.used - origincount)
}

func AppendIfWithinRadius(ga *Array, lon float64, lat float64, radius float64, score float64, member string) int {
	var distance float64
	xy := [2]float64{}

	if !decodehash(score, &xy) {
		return utils.C_ERR
	}
	if !hashGetDistanceIfInRadiusWGS84(lon, lat, xy[0], xy[1], radius, &distance) {
		return utils.C_ERR
	}

	gp := ArrayAppend(ga)
	gp.longitude = xy[0]
	gp.latitude = xy[1]
	gp.dist = distance
	gp.member = member
	gp.score = score
	return utils.C_OK
}

func ArrayAppend(ga *Array) *Point {
	if ga.used == ga.buckets {
		if ga.buckets == 0 {
			ga.buckets = 8
		} else {
			ga.buckets = ga.buckets * 2
		}
	}
	gp := new(Point)
	ga.array = append(ga.array, gp)
	ga.used++
	return gp
}
