package dashboard

import (
	"night-fury/dashboard/api/session"
	wsserver "night-fury/ws_server"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title lisence admin api
// @version 1.0
// @description  lisence server admin api and ws

// @contact.name longalong
// @contact.url https://mastergo.com
// @contact.email huhailong@jwzg.com

//@host 127.0.0.1:8088
func loadRouter(engine *gin.Engine) {

	engine.GET("/ping", func(c *gin.Context) {
		c.String(200, "ok")
	})
	engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := engine.Group("/license/api/v1") // lisence 服务的路径

	apiGroup.Group("/user").
		POST("/signin", session.Signin)

	// ws server
	apiGroup.GET("/hiboss", func(c *gin.Context) {
		wsserver.Serve(c, c.Writer, c.Request)
	})
}
