package date_structure

type Array struct {
	Arr   []*Point
	Buckets uint
	Used    uint
}

type Point struct {
	Longitude float64
	Latitude  float64
	Dist      float64
	Score     float64
	Member    string
}

func ArrayCreate() *Array {
	ga := new(Array)
	ga.Arr = make([]*Point, 0)
	ga.Buckets = 0
	ga.Used = 0
	return ga
}

func ArrayAppend(ga *Array) *Point {
	if ga.Used == ga.Buckets {
		if ga.Buckets == 0 {
			ga.Buckets = 8
		} else {
			ga.Buckets = ga.Buckets * 2
		}
	}
	gp := new(Point)
	ga.Arr = append(ga.Arr, gp)
	ga.Used++
	return gp
}