package config

import (
	"fmt"
)

func init() {
	fmt.Println("config")
}

const (
	DefaultAofFile = "./aof/gkvdb.aof"
)