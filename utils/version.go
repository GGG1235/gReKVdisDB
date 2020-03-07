package utils

import (
	"fmt"
	"os"
)

func ExitHandler() {
	fmt.Println("exiting smoothly ...")
	fmt.Println("bye ")
	os.Exit(0)
}

func Version() {
	println("Gkvdb server v=0.0.1 sha=xxxxxxx:001 malloc=libc-go bits=64 ")
	os.Exit(0)
}

func Usage() {
	println("Usage: ./gkvdb-server [/path/to/redis.conf] [options]")
	println("       ./gkvdb-server - (read config from stdin)")
	println("       ./gkvdb-server -v or --version")
	println("       ./gkvdb-server -h or --help")
	println("Examples:")
	println("       ./gkvdb-server (run the server with default conf)")
	println("       ./gkvdb-server /etc/redis/6379.conf")
	println("       ./gkvdb-server --port 7777")
	println("       ./gkvdb-server --port 7777 --slaveof 127.0.0.1 8888")
	println("       ./gkvdb-server /etc/myredis.conf --loglevel verbose")
	println("Sentinel mode:")
	println("       ./gkvdb-server /etc/sentinel.conf --sentinel")
	os.Exit(0)
}
