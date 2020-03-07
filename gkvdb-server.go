package main

import (
	"fmt"
	"gReKVdisDB/aof"
	"gReKVdisDB/core"
	"gReKVdisDB/lists"
	"gReKVdisDB/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultAofFile = "./aof/gkvdb.aof"
)


var gkvdb = new(core.Server)

func main() {
	
	argv := os.Args
	argc := len(os.Args)
	if argc >= 2 {
		
		if argv[1] == "-v" || argv[1] == "--version" {
			version()
		}
		if argv[1] == "--help" || argv[1] == "-h" {
			usage()
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go sigHandler(c)

	initServer()

	netListen, err := net.Listen("tcp", "0.0.0.0:9736")
	if err != nil {
		log.Print("listen err ")
	}
	
	defer netListen.Close()

	for {
		conn, err := netListen.Accept()

		if err != nil {
			continue
		}
		
		go handle(conn)
	}
}


func handle(conn net.Conn) {
	c := gkvdb.CreateClient()
	for {
		if c.Flags&core.CLIENT_PUBSUB > 0 {
			if c.Buf != "" {
				responseConn(conn, c)
				c.Buf = ""
			}
			time.Sleep(1)

		} else {
			err := c.ReadQueryFromClient(conn)

			if err != nil {
				log.Println("readQueryFromClient err", err)
				return
			}
			err = c.ProcessInputBuffer()
			if err != nil {
				log.Println("ProcessInputBuffer err", err)
				return
			}
			gkvdb.ProcessCommand(c)
			responseConn(conn, c)
		}
	}
}


func responseConn(conn net.Conn, c *core.Client) {
	conn.Write([]byte(c.Buf))
}


func initServer() {
	gkvdb.Pid = os.Getpid()
	gkvdb.DbNum = 16
	initDb()
	gkvdb.Start = time.Now().UnixNano() / 1000000
	
	gkvdb.AofFilename = DefaultAofFile

	getCommand := &core.GkvDBCommand{Name: "get", Proc: core.GetCommand}
	setCommand := &core.GkvDBCommand{Name: "set", Proc: core.SetCommand}
	subscribeCommand := &core.GkvDBCommand{Name: "subscribe", Proc: core.SubscribeCommand}
	publishCommand := &core.GkvDBCommand{Name: "publish", Proc: core.PublishCommand}
	geoaddCommand := &core.GkvDBCommand{Name: "geoadd", Proc: core.GeoAddCommand}
	geohashCommand := &core.GkvDBCommand{Name: "geohash", Proc: core.GeoHashCommand}
	geoposCommand := &core.GkvDBCommand{Name: "geopos", Proc: core.GeoPosCommand}
	geodistCommand := &core.GkvDBCommand{Name: "geodist", Proc: core.GeoDistCommand}
	georadiusCommand := &core.GkvDBCommand{Name: "georadius", Proc: core.GeoRadiusCommand}
	georadiusbymemberCommand := &core.GkvDBCommand{Name: "georadiusbymember", Proc: core.GeoRadiusByMemberCommand}

	gkvdb.Commands = map[string]*core.GkvDBCommand{
		"get":               getCommand,
		"set":               setCommand,
		"geoadd":            geoaddCommand,
		"geohash":           geohashCommand,
		"geopos":            geoposCommand,
		"geodist":           geodistCommand,
		"georadius":         georadiusCommand,
		"georadiusbymember": georadiusbymemberCommand,
		"subscribe":         subscribeCommand,
		"publish":           publishCommand,
	}
	tmp := make(map[string]*lists.List)
	gkvdb.PubSubChannels = &tmp
	LoadData()
}


func initDb() {
	gkvdb.Db = make([]*core.GkvdisDb, gkvdb.DbNum)
	for i := 0; i < gkvdb.DbNum; i++ {
		gkvdb.Db[i] = new(core.GkvdisDb)
		gkvdb.Db[i].Dict = make(map[string]*utils.GKVDBObject, 100)
	}
}
func LoadData() {
	c := gkvdb.CreateClient()
	c.FakeFlag = true
	pros := aof.ReadAof(gkvdb.AofFilename)
	for _, v := range pros {
		c.QueryBuf = string(v)
		err := c.ProcessInputBuffer()
		if err != nil {
			log.Println("ProcessInputBuffer err", err)
		}
		gkvdb.ProcessCommand(c)
	}
}

func sigHandler(c chan os.Signal) {
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitHandler()
		default:
			fmt.Println("signal ", s)
		}
	}
}

func exitHandler() {
	fmt.Println("exiting smoothly ...")
	fmt.Println("bye ")
	os.Exit(0)
}

func version() {
	println("Gkvdb server v=0.0.1 sha=xxxxxxx:001 malloc=libc-go bits=64 ")
	os.Exit(0)
}

func usage() {
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
