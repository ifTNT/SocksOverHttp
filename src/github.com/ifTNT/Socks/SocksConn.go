//Socks connect object
package Socks

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

type SocksConn struct {
	RemoteAddr struct {
		Host string
		Port int
	}
	conn     *net.TCPConn
	writeBuf *bytes.Buffer
	mother   *SocksServer
}

func (s *SocksConn) RemoteAddr2String() string {
	return s.RemoteAddr.Host + ":" + strconv.Itoa(s.RemoteAddr.Port)
}
func (s *SocksConn) SetCanConnect(rep byte) {
	s.conn.Write([]byte{VER, rep, RSV, IPv4, 0, 0, 0, 0, 0, 0})
}

func (s *SocksConn) Write(willWrite []byte) {
	s.writeBuf.Write(willWrite)
}
func (s *SocksConn) Flush() {
	if _, err := s.conn.Write(s.writeBuf.Bytes()); err != nil {
		fmt.Println("Write to " + s.RemoteAddr2String() + " Failed")
		err.Error()
	}
	s.writeBuf.Reset()
}
func (s *SocksConn) Close() {
	s.conn.Close()
}
