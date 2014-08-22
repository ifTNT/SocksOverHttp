// A simple socks server
package Socks

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
)

const (
	BUFFER_SIZE = 1 << 12
	VER         = 0x05
	RSV         = 0x00
	NO_AUTH     = 0x00
	CONNECT     = 0x01
	IPv4        = 0x01
	DOMAINNAME  = 0x03

	iota      = 0x00
	SUCCESSED = iota
)

type SocksServer struct {
	server *net.TCPListener
}

func NewSocksServer(port int) *SocksServer {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("127.0.0.1"), port, ""})
	if err != nil {
		fmt.Printf("Bind with port %d failed\n", port)
		err.Error()
		return nil
	}
	return &SocksServer{ln}
}

type Handle func(*SocksConn, []byte, bool)

func (s *SocksServer) Listen(callback Handle) { // never return
	defer s.Stop()

	for {
		if conn, err := s.server.AcceptTCP(); err != nil {
			fmt.Println("Faild to accept")
			err.Error()
		} else {
			go s.preHandle(conn, make(chan []byte), make(chan error), callback)
		}
	}
}
func asyncRead(conn *net.TCPConn, data chan []byte, ch_err chan error) { //never return
	for {
		buf := make([]byte, BUFFER_SIZE)
		if length, err := conn.Read(buf); err != nil {
			ch_err <- err
			return
		} else {
			buf = buf[:length]
		}
		data <- buf
	}
}
func (s *SocksServer) preHandle(conn *net.TCPConn, ch_data chan []byte, ch_err chan error, callback Handle) {
	go asyncRead(conn, ch_data, ch_err)
	NewSocksConn := SocksConn{}

	for {
		select {
		case err := <-ch_err:
			err.Error()
		case data := <-ch_data:
			fmt.Println("Got some packet")
			fmt.Println("Lenght of packet is " + strconv.Itoa(len(data)) + " Bytes")
			fmt.Println("Content: " + hex.EncodeToString(data))

			if data[0] == byte(VER) {
				if len(data) == int(data[1])+2 {
					conn.Write([]byte{VER, NO_AUTH})
					fmt.Println("Replied")
					continue
				} else if data[1] == byte(CONNECT) {
					switch data[3] {
					case IPv4:
						if len(data) == 10 {
							TempAddr := net.TCPAddr{data[len(data)-6 : len(data)-2], int(data[len(data)-2])<<8 + int(data[len(data)-1]), ""}
							NewSocksConn.RemoteAddr.Host = TempAddr.IP.String()
							NewSocksConn.RemoteAddr.Port = TempAddr.Port
						}
					case DOMAINNAME:
						NewSocksConn.RemoteAddr.Host = bytes.NewBuffer(data[4 : len(data)-2]).String()
						NewSocksConn.RemoteAddr.Port = int(data[len(data)-2]<<8 + data[len(data)-1])
					}
					NewSocksConn.conn = conn
					NewSocksConn.mother = s
					NewSocksConn.writeBuf = bytes.NewBuffer([]byte{})
					fmt.Println("Remote address of socksConn change to " + NewSocksConn.RemoteAddr2String())
					go callback(&NewSocksConn, data, false)
				}
			} else if NewSocksConn.RemoteAddr.Host != "" {
				go callback(&NewSocksConn, data, true)
			} else {
				fmt.Println("Droped illegal packet from " + conn.RemoteAddr().String())
			}
		}

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
