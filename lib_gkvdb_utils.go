package main

import "C"
import (
	"bufio"
	"gReKVdisDB/proto"
	"net"
	"os"
	"strings"
)

func main() {}

//export ReadString
func ReadString(delim byte) string {
	reader := bufio.NewReader(os.Stdin)
	res,err := reader.ReadString(delim)
	if err != nil {
		return ""
	}
	return res
}

//export Replace
func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

//export SendServer
func SendServer(msg string, conn net.Conn) (n int, err error) {
	p, e := proto.EncodeCmd(msg)
	if e != nil {
		return 0, e
	}
	n, err = conn.Write(p)
	return n, err
}

//export DecodeFromBytes
func DecodeFromBytes(p []byte) (*proto.Resp, error) {
	return proto.DecodeFromBytes(p)
}

//export ConnRead
func ConnRead(conn *net.TCPConn,b []byte) (int, error) {
	return conn.Read(b)
}