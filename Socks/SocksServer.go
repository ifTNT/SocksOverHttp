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
)
const ( //Reply
	SUCCEED             = iota
	GEN_SOKCS_FAIL      = iota
	BLOCK_BY_RULES      = iota
	NETWORK_UNREACHABLE = iota
	HOST_UNREACHABLE    = iota
	CONN_REFUSED        = iota
	TTL_EXPIRED         = iota
	UNKNOW_COMMAND      = iota
	ADDR_TYPE_UNSUPPORT = iota
)

type SocksServer struct {
	server *net.TCPListener
}

func NewSocksServer(port int) *SocksServer {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("127.0.0.1"), port, ""})
	if err != nil {
		fmt.Printf("Bind with port %d failure\n", port)
		err.Error()
		return nil
	}
	return &SocksServer{ln}
}

type Handler func(*SocksConn, []byte, bool)

func (s *SocksServer) Listen(callback Handler) { // never return
	defer s.Stop()

	for {
		if conn, err := s.server.AcceptTCP(); err != nil {
			fmt.Println("Failure to accept")
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
			//close(data)
			return
		} else {
			buf = buf[:length]
		}
		data <- buf
	}
}
func (s *SocksServer) preHandle(conn *net.TCPConn, ch_data chan []byte, ch_err chan error, callback Handler) {
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
						} else {
							conn.Write([]byte{VER, ADDR_TYPE_UNSUPPORT, RSV, IPv4, 0, 0, 0, 0, 0, 0})
							conn.Close()
							return
						}
					case DOMAINNAME:
						NewSocksConn.RemoteAddr.Host = bytes.NewBuffer(data[4 : len(data)-2]).String()
						NewSocksConn.RemoteAddr.Port = int(data[len(data)-2]<<8 + data[len(data)-1])
					default:
						conn.Write([]byte{VER, ADDR_TYPE_UNSUPPORT, RSV, IPv4, 0, 0, 0, 0, 0, 0})
						conn.Close()
						return
					}
					NewSocksConn.conn = conn
					NewSocksConn.mother = s
					fmt.Println("Remote address of socksConn change to " + NewSocksConn.RemoteAddr2String())
					go callback(&NewSocksConn, data, false)
				} else {
					conn.Write([]byte{VER, UNKNOW_COMMAND, RSV, IPv4, 0, 0, 0, 0, 0, 0})
					conn.Close()
					return
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
