package wsserver

import (
	"net/http"
	"night-fury/dashboard/api"
	"night-fury/pkgs/utils"
	"night-fury/ws_server/client"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Serve(c *gin.Context, w http.ResponseWriter, r *http.Request) {
	// 创建连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("upgrade ws error : %s", err)
		api.Fail(c, 500, nil)
		return
	}

	clientID := uuid.NewV4().String()
	clientInstance := client.NewClient(conn, clientID)

	err = client.Hub.Register(clientInstance)
	if err != nil {
		utils.RunAfter(func() {
			clientInstance.Close()
		}, time.Second*2, true)

		return
	}

	go clientInstance.ReadMsg()
	go clientInstance.WriteMsg()
}

func connectFail(c *client.Client, msg string) {

	// c.LastMessage()
}
