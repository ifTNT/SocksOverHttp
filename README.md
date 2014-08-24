SocksOverHttp
==========

實作socks5(RFC1928), 然後完全偽裝成http欺騙L7 fliter, Content fliter

##Godoc
###Socks

`import "github.com/YSITD/SocksOverHttp/Socks"`  
>  
> ```go
SocksServer struct{
    server *net.TCPListener	//The real server
}
> ```
>  
> `func NewSocksServer(port int) *SocksServer`  
> Returns a SocksServer that binded on port port.  
>  
> `type Hendler func(*SocksConn)`  
>
> `func (s * SocksServer) Listen(callback Handler)`  
> ***This function will never return***  
> Start listen. If there are any packet that needs user handle, callback will be called.
>
> `func Stop()`  
> Stop the server.  
>  
> ```go
const ( //Reply
	SUCCEED             = iota //Succeed
	GEN_SOKCS_FAIL      = iota //General SOCKS server failure
	BLOCK_BY_RULES      = iota //Connection not allowed by ruleset
	NETWORK_UNREACHABLE = iota //Network unreachable
	HOST_UNREACHABLE    = iota //Host unreachable
	CONN_REFUSED        = iota //Connection refused
	TTL_EXPIRED         = iota //TTL expired
	UNKNOW_COMMAND      = iota //Command not supported
	ADDR_TYPE_UNSUPPORT = iota //Address type not supported
)
> ```  
>  
> ```go
SocksConn struct {
	RemoteAddr struct {
		Host string //The host address that socks client wants to connect
		Port int    //The port that socks client wants to connect
	}
	conn     *net.TCPConn   //The pointer to the real connect object
	writeBuf bytes.Buffer   //Write buffer
	mother   *SocksServer   //Mother server pointer
}
> ```  
>  
> `func (s *SocksConn) RemoteAddr2String() string`  
> Returns the remote address as string.  
>  
> `func (s *SocksConn) SetCanConnect(rep byte)`  
> Tells us whether the remote can connect.  
> The value of rep can only be the const which we provided  
>   
> `func (s *SocksConn) Write(willWrite []byte)`  
> Write some bytes to the buffer. (not the socks client)  
>  
> `func (s *SocksConn) Flush()`  
> Write the buffer to the socks client  
>   
> `func (s *SocksConn) Close()`  
> Close this connect.  
> If you want to continue do anything to this connect during next request, do not close it.

##感謝：

@seadog007  
@ifTNT