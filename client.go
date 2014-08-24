package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/YSITD/SocksOverHttp/Socks"
	"net/http"
)

const (
	SOCKS_PORT  = 23456
	REMOTE_ADDR = "127.0.0.1"
	REMOTE_PORT = "80"
)

func main() {
	server := Socks.NewSocksServer(SOCKS_PORT)
	defer server.Stop()
	server.Listen(func(conn *Socks.SocksConn, data []byte, checked bool) {
		if checked {
			defer conn.Flush()
			fmt.Println("Handling socks request")
			fmt.Println("RemoteAddr: " + conn.RemoteAddr2String())
			fmt.Println("Read bytes: " + hex.EncodeToString(data))
			fmt.Println("Bytes to string :\n" + bytes.NewBuffer(data).String())

			var b64buf, buf bytes.Buffer
			b64 := base64.NewEncoder(base64.StdEncoding, &b64buf)
			b64.Write(data)
			b64.Close()
			fmt.Println("Encoded data: " + b64buf.String())

			buf.Write(bytes.NewBufferString("GIF89a").Bytes()) //head of gif
			gz := gzip.NewWriter(&buf)
			jsonBuf, _ := json.Marshal(map[string]string{
				"dist":    conn.RemoteAddr.Host,
				"port":    string(conn.RemoteAddr.Port),
				"content": b64buf.String()})
			gz.Write(jsonBuf)
			gz.Close()
			buf.Write([]byte{0x3B}) // tail of gif

			fmt.Println("Sendding http request")
			httpClient := &http.Client{}
			req, err := http.NewRequest("GET", "http://"+REMOTE_ADDR+":"+REMOTE_PORT+"/assets/image/XD.gif", &buf)
			if err != nil {
				fmt.Println("Creat http request failure")
				err.Error()
			}

			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/28.0.1469.0 Safari/537.36")
			req.Header.Add("Content-Type", "image/gif")
			if resp, err := httpClient.Do(req); err != nil || resp.StatusCode != 200 {
				fmt.Println("Send http request failure")
				err.Error()
			} else {
				var contentReader, ungzipBuf bytes.Buffer
				contentReader.ReadFrom(resp.Body)
				resp.Body.Close()

				contentBuf := contentReader.Bytes()
				ungzip, err := gzip.NewReader(&ungzipBuf)
				if err != nil {
					fmt.Println("Ungzip failure")
					err.Error()
				}
				ungzip.Read(contentBuf[6 : len(contentBuf)-2])
				ungzip.Close()
				fmt.Println("Writting " + hex.EncodeToString(ungzipBuf.Bytes()) + " to buffer")
				conn.Write(ungzipBuf.Bytes())
			}
			fmt.Println("Handler ended")
		} else {
			fmt.Println("Force return can connect")
			conn.SetCanConnect(Socks.SUCCEED)
		}
	})
}
