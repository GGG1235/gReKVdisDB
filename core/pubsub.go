package core

import (
	"gReKVdisDB/lists"
	"gReKVdisDB/utils"
	"strconv"
)

func SubscribeCommand(c *Client, s *Server) {
	for j := 1; j < c.Argc; j++ {
		pubsubSubscribeChannel(c, c.Argv[j], s)
	}
	c.Flags |= CLIENT_PUBSUB

}

func pubsubSubscribeChannel(c *Client, obj *utils.GKVDBObject, s *Server) {
	(*c.PubSubChannels)[obj.Ptr.(string)] = nil
	de := (*(s.PubSubChannels))[obj.Ptr.(string)]
	var clients *lists.List
	if de == nil {
		clients = lists.ListCreate()
		(*(s.PubSubChannels))[obj.Ptr.(string)] = clients
	} else {
		clients = de
	}
	clients.ListAddNodeTail(c)
}

func PublishCommand(c *Client, s *Server) {
	receivers := pubsubPublishMessage(c.Argv[1], c.Argv[2], s)
	addReplyStatus(c, strconv.Itoa(receivers))
}

func pubsubPublishMessage(channel *utils.GKVDBObject, message *utils.GKVDBObject, s *Server) int {
	receivers := 0
	de := (*s.PubSubChannels)[channel.Ptr.(string)]
	if de != nil {
		for list := de.Head; list != nil; list = list.Next {
			c := list.Value.(*Client)
			addReplyStatus(c, message.Ptr.(string))
			receivers++
		}
	}
	return receivers

}
