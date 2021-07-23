package iclient

type Client interface {
	// 连接相关
	SendMsg([]byte) // 发送消息
	Close()         // 关闭该连接
}
