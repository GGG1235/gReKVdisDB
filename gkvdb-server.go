package main

import (
	"fmt"
	"gReKVdisDB/aof"
	"gReKVdisDB/core"
	"gReKVdisDB/core/date_structure/lists"
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
		
		if argv[1] == "-v" || argv[1] == "--utils" {
			utils.Version()
		}
		if argv[1] == "--help" || argv[1] == "-h" {
			utils.Usage()
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
	addCommand := &core.GkvDBCommand{Name: "add", Proc: core.AddCommand}
	hashCommand := &core.GkvDBCommand{Name: "hash", Proc: core.HashCommand}
	posCommand := &core.GkvDBCommand{Name: "pos", Proc: core.PosCommand}
	distCommand := &core.GkvDBCommand{Name: "dist", Proc: core.DistCommand}
	radiusCommand := &core.GkvDBCommand{Name: "radius", Proc: core.RadiusCommand}
	radiusbymemberCommand := &core.GkvDBCommand{Name: "radiusbymember", Proc: core.RadiusByMemberCommand}

	gkvdb.Commands = map[string]*core.GkvDBCommand{
		"get":               getCommand,
		"set":               setCommand,
		"add":            addCommand,
		"hash":           hashCommand,
		"pos":            posCommand,
		"dist":           distCommand,
		"radius":         radiusCommand,
		"radiusbymember": radiusbymemberCommand,
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
			utils.ExitHandler()
		default:
			fmt.Println("signal ", s)
		}
	}
}