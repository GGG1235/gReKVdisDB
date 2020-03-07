package core

import (
	"fmt"
	"gReKVdisDB/lists"
	"log"
	"os"
)

type Server struct {
	Db               []*GkvdisDb
	DbNum            int
	Start            int64
	Port             int32
	RdbFilename      string
	AofFilename      string
	NextClientID     int32
	SystemMemorySize int32
	Clients          int32
	Pid              int
	Commands         map[string]*GkvDBCommand
	Dirty            int64
	AofBuf           []string
	PubSubChannels   *map[string]*lists.List
	PubSubPatterns   *lists.List
}

func (s *Server) CreateClient() (c *Client) {
	c = new(Client)
	c.Db = s.Db[0]
	c.QueryBuf = ""
	tmp := make(map[string]*lists.List, 0)
	c.PubSubChannels = &tmp
	c.Flags = 0
	return c
}

func (s *Server) ProcessCommand(c *Client) {
	v := c.Argv[0].Ptr
	name, ok := v.(string)
	if !ok {
		log.Println("error cmd")
		os.Exit(1)
	}
	cmd := lookupCommand(name, s)
	fmt.Println(cmd, name, s)
	if cmd != nil {
		c.Cmd = cmd
		call(c, s)
	} else {
		addReplyError(c, fmt.Sprintf("(error) ERR unknown command '%s'", name))
	}
}