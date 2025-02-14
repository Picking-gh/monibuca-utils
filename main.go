package utils

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

// ListenAddrs Listen http and https
func ListenAddrs(addr, addTLS, cert, key string, handler http.Handler) {
	var g errgroup.Group
	if addTLS != "" {
		g.Go(func() error {
			return http.ListenAndServeTLS(addTLS, cert, key, handler)
		})
	}
	if addr != "" {
		g.Go(func() error { return http.ListenAndServe(addr, handler) })
	}
	if err := g.Wait(); err != nil {
		log.Fatal("[utils] ", err)
	}
}

func ListenTCP(addr string, process func(net.Conn)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				Printf("[utils] %s: Accept error: %v; retrying in %v", addr, err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		conn.(*net.TCPConn).SetNoDelay(false)
		tempDelay = 0
		go process(conn)
	}
}

func ListenUDP(address string, networkBuffer int) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("[utils] udp server ResolveUDPAddr :%s error, %v", address, err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("[utils] udp server ListenUDP :%s error, %v", address, err)
	}
	if err = conn.SetReadBuffer(networkBuffer); err != nil {
		Printf("[utils] udp server video conn set read buffer error, %v", err)
	}
	if err = conn.SetWriteBuffer(networkBuffer); err != nil {
		Printf("[utils] udp server video conn set write buffer error, %v", err)
	}
	return conn, err
}

func CORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", origin[0])
	}
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func ReadFileLines(filename string) (lines []string, err error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	bio := bufio.NewReader(file)
	for {
		var line []byte

		line, _, err = bio.ReadLine()
		if err != nil {
			if err == io.EOF {
				file.Close()
				return lines, nil
			}
			return
		}

		lines = append(lines, string(line))
	}
}

func CurrentDir(path ...string) string {
	_, currentFilePath, _, _ := runtime.Caller(1)
	if len(path) == 0 {
		return filepath.Dir(currentFilePath)
	}
	return filepath.Join(filepath.Dir(currentFilePath), filepath.Join(path...))
}
