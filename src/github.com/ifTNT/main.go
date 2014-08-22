package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ifTNT/Socks"
)

func main() {
	server := Socks.NewSocksServer(23456)
	defer server.Stop()
	server.Listen(func(conn *Socks.SocksConn, data []byte, checked bool) {
		if checked {
			defer conn.Flush()
			fmt.Println("Handling socks request")
			fmt.Println("RemoteAddr: " + conn.RemoteAddr2String())
			fmt.Println("Read bytes: " + hex.EncodeToString(data))
			fmt.Println("Bytes to string :\n" + bytes.NewBuffer(data).String())
			fmt.Println("Writting \"abc\" to buffer")
			conn.Write(bytes.NewBufferString("abc").Bytes())
			fmt.Println("Handler ended")
		} else {
			fmt.Println("Force return can connect")
			conn.SetCanConnect(Socks.SUCCEED)
		}
	})
}
