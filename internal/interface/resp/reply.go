package resp

type Reply interface {
	ToBytes() []byte // 将回复转换为字节切片
}

// ErrorReply 错误回复，实现了 Reply 的 ToBytes 方法，也实现了系统的 error 接口
// 这里使用了接口组合，将 error 接口和 Reply 接口组合在一起
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
