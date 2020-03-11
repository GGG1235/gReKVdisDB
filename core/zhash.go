package core

import "math"

const STEP_MAX = 26 /* 26*2 = 52 bits. */

/* Limits from EPSG:900913 / EPSG:3785 / OSGEO:41001 */
const LAT_MIN = -85.05112878
const LAT_MAX = 85.05112878
const LONG_MIN = -180
const LONG_MAX = 180
const D_R = (math.Pi / 180.0)

/// @brief Earth's quatratic mean radius for WGS-84
const EARTH_RADIUS_IN_METERS = 6372797.560856

const MERCATOR_MAX = 20037726.37
const MERCATOR_MIN = -20037726.37

type HashFix52Bits = uint64
type HashVarBits = uint64

type HashBits struct {
	bits uint64
	step uint8
}

type HashRange struct {
	min float64
	max float64
}

type HashArea struct {
	hash      HashBits
	longitude HashRange
	latitude  HashRange
}

type HashRadius struct {
	hash      HashBits
	area      HashArea
	neighbors HashNeighbors
}
type HashNeighbors struct {
	north      HashBits
	east       HashBits
	west       HashBits
	south      HashBits
	north_east HashBits
	south_east HashBits
	north_west HashBits
	south_west HashBits
}

func deg_rad(ang float64) float64 {
	return ang * D_R
}
func rad_deg(ang float64) float64 {
	return ang / D_R
}

func hashEncodeWGS84(longitude float64, latitude float64, step uint8, hash *HashBits) int {
	return hashEncodeType(longitude, latitude, step, hash)
}

func hashEncodeType(longitude float64, latitude float64, step uint8, hash *HashBits) int {
	r := [2]HashRange{}
	hashGetCoordRange(&r[0], &r[1])
	return hashEncode(&r[0], &r[1], longitude, latitude, step, hash)
}

func hashGetCoordRange(long_range *HashRange, lat_range *HashRange) {
	long_range.max = LONG_MAX
	long_range.min = LONG_MIN
	lat_range.max = LAT_MAX
	lat_range.min = LAT_MIN
}

func hashEncode(long_range *HashRange, lat_range *HashRange, longitude float64, latitude float64, step uint8,
	hash *HashBits) int {
	if longitude > 180 || longitude < -180 ||
		latitude > 85.05112878 || latitude < -85.05112878 {
		return 0
	}

	hash.bits = 0
	hash.step = step

	if latitude < lat_range.min || latitude > lat_range.max ||
		longitude < long_range.min || longitude > long_range.max {
		return 0
	}

	var lat_offset float64
	var long_offset float64
	lat_offset =
		(latitude - lat_range.min) / (lat_range.max - lat_range.min)
	long_offset =
		(longitude - long_range.min) / (long_range.max - long_range.min)

	/* convert to fixed point based on the step size */
	mask := 1 << step
	lat_offset = lat_offset * float64(mask)
	long_offset = long_offset * float64(mask)
	hash.bits = interleave64(int32(lat_offset), int32(long_offset))
	return 1
}

func interleave64(latOffset int32, lngOffset int32) uint64 {
	B := []uint64{0x5555555555555555, 0x3333333333333333,
		0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF}
	S := []uint8{1, 2, 4, 8, 16}
	x := uint64(latOffset)
	y := uint64(lngOffset)
	x = (x | (x << S[4])) & B[4]
	y = (y | (y << S[4])) & B[4]
	x = (x | (x << S[3])) & B[3]
	y = (y | (y << S[3])) & B[3]
	x = (x | (x << S[2])) & B[2]
	y = (y | (y << S[2])) & B[2]
	x = (x | (x << S[1])) & B[1]
	y = (y | (y << S[1])) & B[1]
	x = (x | (x << S[0])) & B[0]
	y = (y | (y << S[0])) & B[0]
	return x | (y << 1)
}

func deinterleave64(interleaved uint64) uint64 {
	B := []uint64{0x5555555555555555, 0x3333333333333333,
		0x0F0F0F0F0F0F0F0F, 0x00FF00FF00FF00FF,
		0x0000FFFF0000FFFF, 0x00000000FFFFFFFF}

	S := []uint8{0, 1, 2, 4, 8, 16}
	x := interleaved
	y := interleaved >> 1
	x = (x | (x >> S[0])) & B[0]
	y = (y | (y >> S[0])) & B[0]

	x = (x | (x >> S[1])) & B[1]
	y = (y | (y >> S[1])) & B[1]

	x = (x | (x >> S[2])) & B[2]
	y = (y | (y >> S[2])) & B[2]

	x = (x | (x >> S[3])) & B[3]
	y = (y | (y >> S[3])) & B[3]

	x = (x | (x >> S[4])) & B[4]
	y = (y | (y >> S[4])) & B[4]

	x = (x | (x >> S[5])) & B[5]
	y = (y | (y >> S[5])) & B[5]

	return x | (y << 32)
}

func hashAlign52Bits(hash HashBits) uint64 {
	bits := hash.bits
	bits <<= (52 - hash.step*2)
	return bits
}

func decodehash(bits float64, xy *[2]float64) bool {
	hash := HashBits{bits: uint64(bits), step: STEP_MAX}
	return hashDecodeToLongLatWGS84(hash, xy)
}
func hashDecodeToLongLatWGS84(hash HashBits, xy *[2]float64) bool {
	return hashDecodeToLongLatType(hash, xy)
}

func hashDecodeToLongLatType(hash HashBits, xy *[2]float64) bool {
	area := new(HashArea)
	if xy == nil || !hashDecodeType(hash, area) {
		return false
	}
	return hashDecodeAreaToLongLat(area, xy)
}

func hashDecodeType(hash HashBits, area *HashArea) bool {
	r := [2]HashRange{}
	hashGetCoordRange(&r[0], &r[1])
	return hashDecode(r[0], r[1], hash, area)
}

func hashDecodeWGS84(hash HashBits, area *HashArea) bool {
	return hashDecodeType(hash, area)
}

func hashDecodeAreaToLongLat(area *HashArea, xy *[2]float64) bool {
	if xy == nil {
		return false
	}
	xy[0] = (area.longitude.min + area.longitude.max) / 2
	xy[1] = (area.latitude.min + area.latitude.max) / 2
	return true

}

func hashIsZero(hash HashBits) bool {
	return hash.bits == 0 && hash.step == 0
}

func rangeIsZero(r HashRange) bool {
	return r.max == 0 && r.min == 0
}

func hashDecode(long_range HashRange, lat_range HashRange, hash HashBits, area *HashArea) bool {
	if hashIsZero(hash) || area == nil || rangeIsZero(lat_range) || rangeIsZero(long_range) {
		return false
	}

	area.hash = hash
	step := hash.step
	hash_sep := deinterleave64(hash.bits)

	lat_scale := lat_range.max - lat_range.min
	long_scale := long_range.max - long_range.min

	ilato := uint32(hash_sep)
	ilono := uint32(hash_sep >> 32)

	area.latitude.min = lat_range.min + (float64(ilato)*1.0/float64(uint64(1)<<step))*lat_scale
	area.latitude.max = lat_range.min + ((float64(ilato)+1)*1.0/float64(uint64(1)<<step))*lat_scale
	area.longitude.min = long_range.min + (float64(ilono)*1.0/float64(uint64(1)<<step))*long_scale
	area.longitude.max = long_range.min + ((float64(ilono)+1)*1.0/float64(uint64(1)<<step))*long_scale

	return true
}

func hashGetDistance(lon1d float64, lat1d float64, lon2d float64, lat2d float64) float64 {
	var lat1r, lon1r, lat2r, lon2r, u, v float64
	lat1r = deg_rad(lat1d)
	lon1r = deg_rad(lon1d)
	lat2r = deg_rad(lat2d)
	lon2r = deg_rad(lon2d)
	u = math.Sin((lat2r - lat1r) / 2)
	v = math.Sin((lon2r - lon1r) / 2)
	return 2.0 * EARTH_RADIUS_IN_METERS *
		math.Asin(math.Sqrt(u*u+math.Cos(lat1r)*math.Cos(lat2r)*v*v))
}

func hashGetAreasByRadiusWGS84(longitude float64, latitude float64, radius_meters float64) HashRadius {
	return hashGetAreasByRadius(longitude, latitude, radius_meters)
}

func hashGetAreasByRadius(longitude float64, latitude float64, radius_meters float64) HashRadius {
	var long_range, lat_range HashRange
	var radius HashRadius
	var hash HashBits
	var neighbors HashNeighbors
	var area HashArea
	var min_lon, max_lon, min_lat, max_lat float64
	var bounds [4]float64
	var steps int

	hashBoundingBox(longitude, latitude, radius_meters, &bounds)
	min_lon = bounds[0]
	min_lat = bounds[1]
	max_lon = bounds[2]
	max_lat = bounds[3]

	steps = int(hashEstimateStepsByRadius(radius_meters, latitude))

	hashGetCoordRange(&long_range, &lat_range)                                      //获取经纬度范围 南北极无法code
	hashEncode(&long_range, &lat_range, longitude, latitude, (uint8(steps)), &hash) //hash
	hashNeighbors(&hash, &neighbors)                                                //计算其余8个框的hash
	hashDecode(long_range, lat_range, hash, &area)

	decrease_step := 0
	{
		var north, south, east, west HashArea

		hashDecode(long_range, lat_range, neighbors.north, &north)
		hashDecode(long_range, lat_range, neighbors.south, &south)
		hashDecode(long_range, lat_range, neighbors.east, &east)
		hashDecode(long_range, lat_range, neighbors.west, &west)

		if hashGetDistance(longitude, latitude, longitude, north.latitude.max) < radius_meters {
			decrease_step = 1
		}
		if hashGetDistance(longitude, latitude, longitude, south.latitude.min) < radius_meters {
			decrease_step = 1
		}
		if hashGetDistance(longitude, latitude, east.longitude.max, latitude) < radius_meters {
			decrease_step = 1
		}
		if hashGetDistance(longitude, latitude, west.longitude.min, latitude) < radius_meters {
			decrease_step = 1
		}
	}
	if steps > 1 && decrease_step > 0 {
		steps--
		hashEncode(&long_range, &lat_range, longitude, latitude, uint8(steps), &hash)
		hashNeighbors(&hash, &neighbors)
		hashDecode(long_range, lat_range, hash, &area)
	}
	/* Exclude the search areas that are useless. */
	if steps >= 2 {
		if area.latitude.min < min_lat {
			GZERO(&neighbors.south)
			GZERO(&neighbors.south_west)
			GZERO(&neighbors.south_east)
		}
		if area.latitude.max > max_lat {
			GZERO(&neighbors.north)
			GZERO(&neighbors.north_east)
			GZERO(&neighbors.north_west)
		}
		if area.longitude.min < min_lon {
			GZERO(&neighbors.west)
			GZERO(&neighbors.south_west)
			GZERO(&neighbors.north_west)
		}
		if area.longitude.max > max_lon {
			GZERO(&neighbors.east)
			GZERO(&neighbors.south_east)
			GZERO(&neighbors.north_east)
		}
	}
	radius.hash = hash
	radius.neighbors = neighbors
	radius.area = area
	return radius
}
func GZERO(s *HashBits) {
	s.bits = 0
	s.step = 0
}

//计算经度、纬度为中心的搜索区域的边界框
func hashBoundingBox(longitude float64, latitude float64, radius_meters float64, bounds *[4]float64) bool {
	if bounds == nil {
		return false
	}
	bounds[0] = longitude - rad_deg(radius_meters/EARTH_RADIUS_IN_METERS/math.Cos(deg_rad(latitude)))
	bounds[2] = longitude + rad_deg(radius_meters/EARTH_RADIUS_IN_METERS/math.Cos(deg_rad(latitude)))
	bounds[1] = latitude - rad_deg(radius_meters/EARTH_RADIUS_IN_METERS)
	bounds[3] = latitude + rad_deg(radius_meters/EARTH_RADIUS_IN_METERS)
	return true
}

//计算bits 位的精度
func hashEstimateStepsByRadius(range_meters float64, lat float64) uint8 {
	if range_meters == 0 {
		return 26
	}
	step := uint8(1)
	for range_meters < MERCATOR_MAX {
		range_meters *= 2
		step++
	}

	step -= 2
	if lat > 66 || lat < -66 {
		step--
		if lat > 80 || lat < -80 {
			step--
		}
	}

	/* Frame to valid range. */
	if step < 1 {
		step = 1
	}
	if step > 26 {
		step = 26
	}
	return step
}

//计算其余8个框的hash
func hashNeighbors(hash *HashBits, neighbors *HashNeighbors) {
	neighbors.east = *hash
	neighbors.west = *hash
	neighbors.north = *hash
	neighbors.south = *hash
	neighbors.south_east = *hash
	neighbors.south_west = *hash
	neighbors.north_east = *hash
	neighbors.north_west = *hash //8个方位的hash赋值

	hash_move_x(&neighbors.east, 1)
	hash_move_y(&neighbors.east, 0)

	hash_move_x(&neighbors.west, -1)
	hash_move_y(&neighbors.west, 0)

	hash_move_x(&neighbors.south, 0)
	hash_move_y(&neighbors.south, -1)

	hash_move_x(&neighbors.north, 0)
	hash_move_y(&neighbors.north, 1)

	hash_move_x(&neighbors.north_west, -1)
	hash_move_y(&neighbors.north_west, 1)

	hash_move_x(&neighbors.north_east, 1)
	hash_move_y(&neighbors.north_east, 1)

	hash_move_x(&neighbors.south_east, 1)
	hash_move_y(&neighbors.south_east, -1)

	hash_move_x(&neighbors.south_west, -1)
	hash_move_y(&neighbors.south_west, -1)
}

func hash_move_x(hash *HashBits, d int8) {
	if d == 0 {
		return
	}

	x := hash.bits & 0xaaaaaaaaaaaaaaaa
	y := hash.bits & 0x5555555555555555

	zz := uint64(0x5555555555555555 >> (64 - hash.step*2))
	if d > 0 {
		x = x + (zz + 1)
	} else {
		x = x | zz
		x = x - (zz + 1)
	}
	x &= (0xaaaaaaaaaaaaaaaa >> (64 - hash.step*2))
	hash.bits = (x | y)
}

func hash_move_y(hash *HashBits, d int8) {
	if d == 0 {
		return
	}

	x := hash.bits & 0xaaaaaaaaaaaaaaaa
	y := hash.bits & 0x5555555555555555

	zz := uint64(0xaaaaaaaaaaaaaaaa >> (64 - hash.step*2))
	if d > 0 {
		y = y + (zz + 1)
	} else {
		y = y | zz
		y = y - (zz + 1)
	}
	y &= (0x5555555555555555 >> (64 - hash.step*2))
	hash.bits = (x | y)
}

func hashGetDistanceIfInRadius(x1 float64, y1 float64, x2 float64, y2 float64, radius float64, distance *float64) bool {
	*distance = hashGetDistance(x1, y1, x2, y2)
	if *distance > radius {
		return false
	}
	return true
}

func hashGetDistanceIfInRadiusWGS84(x1 float64, y1 float64, x2 float64, y2 float64, radius float64, distance *float64) bool {
	return hashGetDistanceIfInRadius(x1, y1, x2, y2, radius, distance)
}
