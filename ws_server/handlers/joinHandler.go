package handlers

import (
	"context"
	"night-fury/pkgs/log"
	client "night-fury/ws_server/iclient"
)

type paramJoin struct {
	ID        string `json:"ID"`
	SecretKey string `json:"secretKey"`
}

func HandleJoin(ctx context.Context, c client.Client, msgType int, msg []byte) {
	param := &paramJoin{}
	err := decodeBindMetaData(msg, param)
	if err != nil {
		log.Errorf(log.TagActionJoin, "decode parameter error %s", err)
		return
	}

	// 做一些处理
}
