package client

import (
	"context"
	"night-fury/pkgs/log"
	"night-fury/pkgs/utils"
	"night-fury/ws_server/handlers"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID string

	closeChan chan struct{}
	conn      *websocket.Conn
	msgChan   chan []byte

	hub *ClientHub

	lastPingTime *time.Time

	closingFlag bool
}

func NewClient(conn *websocket.Conn, ID string) *Client {
	now := time.Now()

	return &Client{
		ID:           ID,
		conn:         conn,
		closeChan:    make(chan struct{}, 1),
		msgChan:      make(chan []byte, 60),
		lastPingTime: &now,
	}
}

func (c *Client) SendMsg(msg []byte) {
	c.msgChan <- msg
}

func (c *Client) ReadMsg() {
	defer func() {
		c.Close()
	}()
	c.conn.SetReadLimit(1024 * 1024 * 50) // 50 mb

	for !c.closingFlag {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Warnf(log.TagWSServer, "read msg error: %v", err)
			} else {
				log.Errorf(log.TagWSServer, "user close the client : %s", err)
			}
			break
		}
		go utils.SafeRun(nil, func() {
			c.handleMsg(message)
		})
	}
}

func (c *Client) WriteMsg() {
	heartBeatTicker := time.NewTicker(time.Second * 30)
	pingTicker := time.NewTicker(time.Second * 30)
	defer func() {
		heartBeatTicker.Stop()
		pingTicker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.msgChan:
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.BinaryMessage, message)

			if err != nil {
				return
			}

		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			c.conn.WriteMessage(websocket.PingMessage, nil)

		case <-heartBeatTicker.C:
			// 心跳检查
			if c.lastPingTime.Add(time.Minute).Before(time.Now()) {
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

func (c *Client) handleMsg(msg []byte) {
	// 处理message类型，并进行分发
	msgType, data := handlers.DecodeMsgType(msg)

	// 鉴权，是否能够发送该类型的消息
	if !handlers.IsUserMsgType(msgType) {
		log.Errorln(log.TagWSServer, "messagetype not allowed")
		return
	}

	// 获取处理器并处理
	handlerFunc, err := handlers.MessageHandlers.GetHandler(msgType)
	if err != nil {
		log.Errorf(log.TagWSServer, "get msg handler error : %s", err)
	}
	msgCtx := &handlers.MsgContext{
		Ctx:    context.Background(),
		Client: c,
		Msg:    data,
	}
	handlerFunc(msgCtx)
}

func (c *Client) LastMessage(msg []byte) {
	c.conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second*5))

	err := utils.RunAfter(func() {
		c.Close()
	}, time.Second*5, true)
	if err != nil {
		log.Errorf(log.TagWSServer, "close client error : %s", err)
	}
}

func (c *Client) Close() {
	if c.closingFlag {
		return
	}
	var err error
	c.closingFlag = true

	if err = c.conn.Close(); err != nil {
		log.Errorf(log.TagWSServer, "close client error : %s", err)
	}

	err = utils.RunAfter(func() {
		c.hub.UnRegister(c.ID)
	}, time.Second*5, true)
	if err != nil {
		log.Errorf(log.TagWSServer, "close client error : %s", err)
	}
}
