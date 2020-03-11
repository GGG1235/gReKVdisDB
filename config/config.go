package config

import (
	"fmt"
	"sync"
)

func init() {
	var config Config
	config.once.Do(func (){
		fmt.Println("config")
	})
}

type Config struct {
	once sync.Once
}