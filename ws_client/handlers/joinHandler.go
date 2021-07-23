package handlers

import (
	"context"
	"night-fury/pkgs/log"
	"night-fury/ws_client/client"
	"night-fury/ws_client/enum"
)

type ResJoinParam struct {
	Succ bool `json:"succ"`
}

func handleJoin(_ context.Context, c *client.Client, msgType int, msg []byte) {
	p := &ResJoinParam{}
	err := decodeBindMetaData(msg, p)
	if err != nil {
		log.Errorf(log.TagWSClient, "decode bind meta data error %s", err)
	}

}

type connectParam struct {
	ID string `json:"ID"`
}

func connectToServer(c *client.Client) error {
	msg := &connectParam{
		ID: "hello world",
	}

	b, err := EncodeJSON(enum.TYPE_JOIN, msg)
	if err != nil {
		return err
	}
	c.SendMsg(b)
	return nil
}
