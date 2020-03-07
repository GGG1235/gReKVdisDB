package main

import (
	"bufio"
	"fmt"
	"gReKVdisDB/proto"
	"gReKVdisDB/utils"
	"net"
	"os"
	"strings"
)

func main() {

	var IPPort string
	argv := os.Args
	argc := len(argv)

	if argc == 2 {
		IPPort = argv[1]
	} else {
		IPPort = "127.0.0.1:9736"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Hi GkvDB")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", IPPort)
	utils.CheckError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	utils.CheckError(err)
	defer conn.Close()

	for {
		fmt.Print(IPPort + "> ")
		text, _ := reader.ReadString('\n')
		
		text = strings.Replace(text, "\n", "", -1)
		send2Server(text, conn)

		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		resp, er := proto.DecodeFromBytes(buff)
		utils.CheckError(err)
		if n == 0 {
			fmt.Println(IPPort+"> ", "nil")
		} else if er == nil {
			fmt.Println(IPPort+">", string(resp.Value))
		} else {
			fmt.Println(IPPort+"> ", "err server response")
		}
	}

}
func send2Server(msg string, conn net.Conn) (n int, err error) {
	p, e := proto.EncodeCmd(msg)
	if e != nil {
		return 0, e
	}
	
	n, err = conn.Write(p)
	return n, err
}