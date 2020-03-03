package proto

import (
	"gReKVdisDB/utils"
	"strconv"
)

const (
	MaxBulkBytesLen = 1024 * 1024 * 512
	MaxArrayLen = 1024 * 1024
)

type RespType byte

const (
	TypeString    = '+'
	TypeError     = '-'
	TypeInt       = ':'
	TypeBulkBytes = '$'
	TypeArray     = '*'
)

func Btoi64(b []byte) (int64, error) {
	if len(b) != 0 && len(b) < 10 {
		var neg, i = false, 0
		switch b[0] {
		case '-':
			neg = true
			fallthrough
		case '+':
			i++
		}
		if len(b) != i {
			var n int64
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = int64(b[i]-'0') + n*10
			}
			if len(b) == i {
				if neg {
					n = -n
				}
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseInt(string(b), 10, 64); err != nil {
		return 0, utils.ErrorsTrace(err)
	} else {
		return n, nil
	}
}