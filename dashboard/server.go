package dashboard

import (
	"fmt"
	"night-fury/dashboard/intercepter"
	"night-fury/pkgs/log"
	"os"

	"github.com/gin-gonic/gin"
	"gitlab.lanhuapp.com/gopkgs/config"
)

type Server struct {
	engine *gin.Engine
}

func NewServer() *Server {
	var engine *gin.Engine

	switch os.Getenv("ENV") {
	case "local":
		engine = gin.Default()
		gin.SetMode(gin.DebugMode)
	default:
		engine = gin.New()
		engine.Use(gin.Recovery())
		gin.SetMode(gin.ReleaseMode)
	}

	engine.Use(intercepter.MiddleWareCors())
	engine.Use(intercepter.MiddleWareLog())

	loadRouter(engine)

	return &Server{
		engine: engine,
	}
}

func (s *Server) Serve() {
	host := config.GetString("server.host")
	port := config.GetInt64("server.port")

	addr := fmt.Sprintf("%s:%d", host, port)

	log.Infof(log.TagServer, "HTTP server listening on %s", addr)

	log.Fatalf(log.TagServer, "HTTP server start error : %s", s.engine.Run(addr))
}
