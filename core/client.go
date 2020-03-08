package core

import (
	"bytes"
	"errors"
	"gReKVdisDB/aof"
	"gReKVdisDB/date_structure/lists"
	"gReKVdisDB/proto"
	"gReKVdisDB/utils"
	"log"
	"net"
)

type Client struct {
	Cmd            *GkvDBCommand
	Argv           []*utils.GKVDBObject
	Argc           int
	Db             *GkvdisDb
	QueryBuf       string
	Buf            string
	FakeFlag       bool
	PubSubChannels *map[string]*lists.List
	PubSubPatterns *lists.List
	Flags          int 
}

const CLIENT_PUBSUB = (1 << 18)

type Dict map[string]*utils.GKVDBObject

func SetCommand(c *Client, s *Server) {
	objKey := c.Argv[1]
	objValue := c.Argv[2]
	if c.Argc != 3 {
		addReplyError(c, "(error) ERR wrong number of arguments for 'set' command")
	}
	if stringKey, ok1 := objKey.Ptr.(string); ok1 {
		if stringValue, ok2 := objValue.Ptr.(string); ok2 {
			c.Db.Dict[stringKey] = utils.CreateObject(utils.ObjectTypeString, stringValue)
		}
	}
	s.Dirty++
	addReplyStatus(c, "OK")
}


func GetCommand(c *Client, s *Server) {
	o := LookupKey(c.Db, c.Argv[1])
	if o != nil {
		addReplyStatus(c, o.Ptr.(string))
	} else {
		addReplyStatus(c, "nil")
	}
}


func addReply(c *Client, o *utils.GKVDBObject) {
	c.Buf = o.Ptr.(string)
}

func addReplyStatus(c *Client, s string) {
	r := proto.NewString([]byte(s))
	addReplyString(c, r)
}
func addReplyError(c *Client, s string) {
	r := proto.NewError([]byte(s))
	addReplyString(c, r)
}
func addReplyString(c *Client, r *proto.Resp) {
	if ret, err := proto.EncodeToBytes(r); err == nil {
		c.Buf = string(ret)
	}
}


func call(c *Client, s *Server) {
	dirty := s.Dirty
	c.Cmd.Proc(c, s)
	dirty = s.Dirty - dirty
	if dirty > 0 && !c.FakeFlag {
		aof.AppendToFile(s.AofFilename, c.QueryBuf)
	}

}

func (c *Client) ReadQueryFromClient(conn net.Conn) (err error) {
	buff := make([]byte, 512)
	n, err := conn.Read(buff)

	if err != nil {
		log.Println("conn.Read err!=nil", err, "---len---", n, conn)
		conn.Close()
		return err
	}
	c.QueryBuf = string(buff)
	return nil
}


func (c *Client) ProcessInputBuffer() error {
	decoder := proto.NewDecoder(bytes.NewReader([]byte(c.QueryBuf)))
	if resp, err := decoder.DecodeMultiBulk(); err == nil {
		c.Argc = len(resp)
		c.Argv = make([]*utils.GKVDBObject, c.Argc)
		for k, s := range resp {
			c.Argv[k] = utils.CreateObject(utils.ObjectTypeString, string(s.Value))
		}
		return nil
	}
	return errors.New("ProcessInputBuffer failed")
}