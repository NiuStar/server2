package server2

import (
	"crypto/tls"
	"fmt"
	"github.com/NiuStar/log"
	"github.com/NiuStar/utils"
	"github.com/NiuStar/server2/STNet"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/websocket"
	xlog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"github.com/DeanThompson/ginpprof"
	xgin "github.com/NiuStar/server2/gin"
	"github.com/NiuStar/Basic"
)



type XServer struct {

	address string

	server   *gin.Engine
	listener net.Listener

	signalChan      chan os.Signal
	downloadHandler func(filename, username, ip string, fullPath string)

	getRoutes    map[string]xgin.HandlerFunc
	postRoutes   map[string]xgin.HandlerFunc
	socketRoutes map[string]websocket.Handler

	mainRouter   string
	w *xgin.WebSocketServices
}

func Default() *XServer {
	ser := NewServer(":", DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT)
	ser.getRoutes = make(map[string]xgin.HandlerFunc)
	ser.postRoutes = make(map[string]xgin.HandlerFunc)
	ser.socketRoutes = make(map[string]websocket.Handler)

	ser.w = xgin.InitSocketService()

	ser.w.AddPOSTService("RequestRouter",STNet.RequestRouter)
	ser.address = ":" + Basic.GetServerConfig().ServerConfig.GetPort()
	return ser
}

func (s *XServer)SetMainRouter(mainRouter string) {
	s.mainRouter = mainRouter
	s.POST(mainRouter,STNet.RequestRouter)
}

func (s *XServer) GET(key string, call xgin.HandlerFunc) {

	s.w.AddGETService(key, call)
	s.getRoutes[key] = call

}
func (s *XServer) POST(key string, call xgin.HandlerFunc) {

	fmt.Println("POST方法")
	s.w.AddPOSTService(key, call)
	s.postRoutes[key] = call

}

func (s *XServer) HandfuncWebSocket(key string, call websocket.Handler) {
	s.socketRoutes[key] = call
}

func (s *XServer) DownloadFileDelegate(call func(string, string, string, string)) {
	s.downloadHandler = call
}

func (this *XServer) RunServer() {

	if Basic.GetServerConfig().LogConfig != nil {
		log.SetSaveDays(int(Basic.GetServerConfig().LogConfig.Log_days))
	}
	gin.SetMode(gin.DebugMode)

	for key, value := range this.getRoutes {
		this.server.GET(key, CreateGinConetxt(value))
	}
	for key, value := range this.postRoutes {

		this.server.POST(key, CreateGinConetxt(value))
	}

	if Basic.GetServerConfig().ServerConfig.Websocket {
		this.HandfuncWebSocket("/ws", websocket.Handler(this.w.EchoHandler))

		for key, value := range this.socketRoutes {
			s := new(websocket.Server)
			s.Handler = value
			this.server.GET(key, func(c *gin.Context) {
				s.ServeHTTP(c.Writer, c.Request)
			})
		}
	}


	this.initFileServer()
	ginpprof.Wrapper(this.server)

	if Basic.GetServerConfig().ServerConfig.TLS {

		this.ListenAndServerTLS(utils.GetCurrPath()+Basic.GetServerConfig().ServerConfig.Cert, Basic.GetServerConfig().ServerConfig.Key)
	} else {
		this.ListenAndServer()
	}
}

func CreateGinConetxt(value xgin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xgin.NewContext(c)
		value(xc)
	}

}

const (
	GRACEFUL_ENVIRON_KEY    = "IS_GRACEFUL"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_READ_TIMEOUT  = 60 * time.Second
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
)

// new server
func NewServer(addr string, readTimeout, writeTimeout time.Duration) *XServer {
	fmt.Println("port=", addr)
	log.Init()

	// 获取环境变量
//	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		//isGraceful = true
	}

	// 实例化Server
	return &XServer{
		address:           addr,
		server:            gin.Default(),
		//isGraceful:        isGraceful,
		//HadAllowAllMethod: false,
		signalChan:        make(chan os.Signal),
	}
}

func (this *XServer) ListenAndServer() error {
	return this.Server()
}

func (this *XServer) ListenAndServerTLS(certFile, keyFile string) error {

	config := &tls.Config{}
	srv, _ := this.server.GetTLSConfig(this.address)
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
		return err
	}

	ln, err := this.server.GetTCPListener(false, this.address)
	if err != nil {
		panic(err)
		return err
	}

	this.listener = newListener(ln.(*net.TCPListener))
	return this.ServerTLS(srv, tls.NewListener(this.listener, config))
}

func (this *XServer) Server() error {

	xlog.Println(fmt.Sprintf("(this *Server) Serve() with pid %d.  "+this.address, os.Getpid()))
	ln, err := this.server.GetTCPListener(false, this.address)
	if err != nil {
		xlog.Println(fmt.Sprintf("1a56 with pid %d.", os.Getpid()))
		fmt.Println("1a56")
		panic(err)
		return err
	}
	this.listener = newListener(ln.(*net.TCPListener))

	this.server.Run(this.listener, this.address)

	// 跳出Serve处理代表 listener 已经close，等待所有已有的连接处理结束
	this.logf("waiting for connection close...")
	this.listener.(*Listener).Wait()
	this.logf("all connection closed, process with pid %d shutting down...", os.Getpid())

	return err
}

func (this *XServer) ServerTLS(srv http.Server, ln net.Listener) error {

	this.server.RunTLS(srv, ln)

	// 跳出Serve处理代表 listener 已经close，等待所有已有的连接处理结束
	this.logf("waiting for connection close...")
	this.listener.(*Listener).Wait()
	this.logf("all connection closed, process with pid %d shutting down...", os.Getpid())

	return nil
}

func GetFullPath(path string) string {
	absolutePath, _ := filepath.Abs(path)
	return absolutePath
}

func GetTimeFromList(name string, l []interface{}) int64 {
	for _, value := range l {
		list := value.(map[string]interface{})

		if name == list["name"].(string) {
			return list["modtime"].(int64)
		}
	}
	return 0
}

func (this *XServer) logf(format string, args ...interface{}) {

	xlog.Printf(format, args...)

}
