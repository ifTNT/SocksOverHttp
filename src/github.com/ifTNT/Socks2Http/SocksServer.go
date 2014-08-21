// A simple socks server
package Socks2Http

import (
	"bytes"
	"fmt"
	"net"
)

const (
	BUFFER_SIZE = 2 ^ 10
	VER         = 0x05
	RSV         = 0x00
	NO_AUTH     = 0x00
	CONNECT     = 0x01
	IPv4        = 0x01
	DOMAINNAME  = 0x03
)

type SocksServer struct {
	server  *net.TCPListener
	connSet map[string]*SocksConn
}

func NewSocksServer(port int) *SocksServer {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("127.0.0.1"), port, ""})
	if err != nil {
		fmt.Printf("Bind with port %d failed", port)
		err.Error()
		return nil
	}
	return &SocksServer{ln, make(map[string]*SocksConn)}
}

type Handle func(*SocksConn)

func (s *SocksServer) Listen(callback Handle) { // never return
	defer s.Stop()

	for {
		if conn, err := s.server.AcceptTCP(); err != nil {
			fmt.Println("Faild to accept")
			err.Error()
		} else {
			go s.preHandle(conn, callback)
		}
	}
}
func (s *SocksServer) preHandle(conn *net.TCPConn, callback Handle) {
	buf := make([]byte, BUFFER_SIZE)

	if _, err := conn.Read(buf); err != nil {
		fmt.Println("Faild to read buffer")
		err.Error()
	}
	if bytes.Contains(buf[:1], []byte{VER}) {
		if len(buf) == int(buf[1])+2 {
			conn.Write([]byte{VER, NO_AUTH})
		}
		if buf[1] == CONNECT {
			NewSocksConn := SocksConn{}
			switch {
			case buf[3] == IPv4 && len(buf) == 10:
				TempAddr := net.TCPAddr{buf[len(buf)-6 : len(buf)-2], int(buf[len(buf)-2]<<8 + buf[len(buf)-1]), ""}
				NewSocksConn.RemoteAddr.Host = TempAddr.IP.String()
				NewSocksConn.RemoteAddr.Port = TempAddr.Port
			case buf[3] == DOMAINNAME:
				NewSocksConn.RemoteAddr.Host = bytes.NewBuffer(buf[4 : len(buf)-2]).String()
				NewSocksConn.RemoteAddr.Port = int(buf[len(buf)-2]<<8 + buf[len(buf)-1])
			}
			NewSocksConn.conn = conn
			NewSocksConn.mother = s
			NewSocksConn.writeBuf = bytes.NewBuffer(make([]byte, BUFFER_SIZE))
			s.connSet[conn.RemoteAddr().String()] = &NewSocksConn
		}
	}
	if SocksConnRegistered, found := s.connSet[conn.RemoteAddr().String()]; found {
		go callback(SocksConnRegistered)
	} else {
		fmt.Println("Droped illegal packet from " + conn.RemoteAddr().String())
	}
}
func (s *SocksServer) Stop() {
	s.server.Close()
}

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
	return s.RemoteAddr.Host + ":" + string(s.RemoteAddr.Port)
}
func (s *SocksConn) SetCanConnect(status byte) {
	s.conn.Write([]byte{VER, status, RSV, IPv4, 0, 0, 0, 0, 0, 0})
}
func (s *SocksConn) Read(readBytes []byte) {
	if _, err := s.conn.Read(readBytes); err != nil {
		fmt.Println("Read from " + s.RemoteAddr2String() + " Failed")
		err.Error()
	}
}
func (s *SocksConn) Write(willWrite []byte) {
	s.writeBuf.Write(willWrite)
}
func (s *SocksConn) FlushAndClose() {
	if _, err := s.conn.Write(s.writeBuf.Bytes()); err != nil {
		fmt.Println("Write to " + s.RemoteAddr2String() + " Failed")
		err.Error()
	}
	delete(s.mother.connSet, s.RemoteAddr2String())
	s.conn.Close()
}
