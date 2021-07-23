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

	conn *websocket.Conn

	msgChan     chan []byte
	closeChan   chan struct{}
	closingFlag bool

	lastPongTime *time.Time
}

func NewClient(addr string, closeChan chan struct{}) (*Client, error) {
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	t := time.Now()
	client := &Client{
		conn:      c,
		msgChan:   make(chan []byte, 50),
		closeChan: closeChan,

		lastPongTime: &t,
	}

	// 设置大小限制
	client.conn.SetReadLimit(1024 * 1024 * 50) // 50 mb

	// 设置pong处理器
	client.conn.SetPongHandler(func(appData string) error {
		t := time.Now()
		client.lastPongTime = &t
		return nil
	})
	// 设置关闭处理器
	client.conn.SetCloseHandler(func(code int, text string) error {
		client.closeChan <- struct{}{}
		return nil
	})

	return client, nil
}

func (c *Client) SendMsg(msg []byte) {
	c.msgChan <- msg
}

func (c *Client) ReadMsg() {
	defer func() {
		c.Close()
	}()

	for !c.closingFlag {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Warnf(log.TagWSClient, "read msg error: %v", err)
			} else {
				log.Errorf(log.TagWSClient, "server close the client : %s", err)
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
			if c.lastPongTime.Add(time.Minute).Before(time.Now()) {
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
		log.Errorln(log.TagWSClient, "messagetype not allowed")
		return
	}

	// 获取处理器并处理
	handlerFunc, err := handlers.MessageHandlers.GetHandler(msgType)
	if err != nil {
		log.Errorf(log.TagWSClient, "get msg handler error : %s", err)
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
		log.Errorf(log.TagWSClient, "close client error : %s", err)
	}
}

func (c *Client) Close() {
	if c.closingFlag {
		return
	}
	var err error
	c.closingFlag = true

	if err = c.conn.Close(); err != nil {
		log.Warnf(log.TagWSClient, "close client error : %s", err)
	}

}
