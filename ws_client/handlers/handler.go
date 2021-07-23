package handlers

import (
	"context"
	"encoding/binary"
	"errors"
	"night-fury/pkgs/log"
	"night-fury/ws_client/client"
	"night-fury/ws_client/enum"

	jsoniter "github.com/json-iterator/go"
)

var MessageHandlers msgHandlerHub

var ErrNoHandler = errors.New("no such handler")

func init() {
	// 错误处理消息
	MessageHandlers.RegisterHandler(newHandler(enum.TYPE_ERR_MSG, handleErrMsgType))

}

type msgHandlerHub struct {
	msgHandlers map[int]MessageHandler
}

type MsgContext struct {
	Ctx    context.Context
	Client *client.Client
	Msg    []byte
}

func (m *msgHandlerHub) RegisterHandler(handler MessageHandler) {
	handlerID := handler.GetID()
	m.msgHandlers[handlerID] = handler

	if err := handler.Init(); err != nil {
		log.NS().Panicf("init handler error %s", err)
	}
}

func (m *msgHandlerHub) GetHandler(msgType int) (func(*MsgContext), error) {
	handler, ok := m.msgHandlers[msgType]
	if !ok {
		return m.msgHandlers[enum.TYPE_ERR_MSG].Handle, ErrNoHandler
	}
	return handler.Handle, nil
}

type MessageHandler interface {
	GetID() int
	Handle(*MsgContext)
	Init() error
}

func newHandler(msgTypeID int, handler func(context.Context, *client.Client, int, []byte)) MessageHandler {
	return &baseHandler{
		id:      msgTypeID,
		handler: handler,
		msgChan: make(chan *MsgContext, 30),
	}
}

type baseHandler struct {
	id      int
	handler func(context.Context, *client.Client, int, []byte)
	msgChan chan *MsgContext
}

func (h *baseHandler) GetID() int {
	return h.id
}
func (h *baseHandler) Handle(msgCtx *MsgContext) {
	h.msgChan <- msgCtx
}
func (h *baseHandler) Init() error {
	go func() {
		var msg *MsgContext
		for msg = range h.msgChan {
			h.handler(msg.Ctx, msg.Client, h.id, msg.Msg)
		}
	}()
	return nil
}

func formatMsgType(msgType int) []byte {
	tb := make([]byte, 2) // 消息类型最大 65535
	binary.BigEndian.PutUint16(tb, uint16(msgType))
	return tb
}
func formatMsgLen(msgLen int) []byte {
	tb := make([]byte, 4)
	binary.BigEndian.PutUint32(tb, uint32(msgLen))
	return tb
}

func handleErrMsgType(_ context.Context, c *client.Client, msgType int, msg []byte) {
	errMsg := map[string]interface{}{
		"code":    enum.CODE_ERR_NO_MSGTYPE,
		"message": "no such message type",
	}

	msgByte, err := EncodeJSON(msgType, errMsg)
	if err != nil {
		log.Errorf(log.TagWS, "encode json error : %s", err)
		return
	}
	c.SendMsg(msgByte)
}

func EncodeNil(msgType int) []byte {
	b := formatMsgType(msgType)
	b = append(b, formatMsgLen(0)...)
	return b
}

func EncodeJSON(msgType int, jsonData interface{}) ([]byte, error) {
	b := formatMsgType(msgType)
	if jsonData == nil {
		jsonData = &struct{}{}
	}
	byteData, err := jsoniter.Marshal(jsonData)
	if err != nil {
		return nil, err
	}
	b = append(b, formatMsgLen(len(byteData))...)
	b = append(b, byteData...)
	return b, nil
}

func decodeBindMetaData(data []byte, pointer interface{}) error {
	metaLen := binary.BigEndian.Uint32(data[0:4])

	if len(data) < int(metaLen)+4 {
		return errors.New("meta data decode error: no enough data")
	}
	metaBytes := data[4 : metaLen+4]

	return jsoniter.Unmarshal(metaBytes, pointer)
}

func DecodeMsgType(data []byte) (int, []byte) {
	if len(data) < 2 {
		// 消息格式错误
		return enum.TYPE_ERR_MSG, []byte{}
	}
	msgTypeByte := data[0:2]
	if len(data) == 2 {
		return int(binary.BigEndian.Uint16(msgTypeByte)), []byte{}
	}
	return int(binary.BigEndian.Uint16(msgTypeByte)), data[2:]
}
