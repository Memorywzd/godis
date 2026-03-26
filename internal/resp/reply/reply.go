package reply

import (
	"bytes"
	"godis/internal/interface/resp"
	"strconv"
)

var (
	nullBUlkReplyBytes = []byte("$-1") // -1，表示 nil 值
	CRLF               = "\r\n"
)

// BulkReply 字符串回复
type BulkReply struct {
	Arg []byte // 回复的内容，此时是不符合 RESP 协议的
}

func (r *BulkReply) ToBytes() []byte {
	// 如果字符串为空，返回空字符串
	if len(r.Arg) == 0 {
		return nullBUlkReplyBytes
	}
	// 将 BulkReply 转换为符合 RESP 协议的字节数组
	// TODO: 计算[]byte长度的方法
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// MultiBulkReply 多个字符串回复
type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			// *-1\r\n\r\n 表示空数组
			buf.WriteString(string(nullBUlkReplyBytes) + CRLF)
		} else {
			// *3\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$5\r\nhello\r\n
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	// 返回的内容是一个字节切片
	return buf.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// StandardErrorReply 状态回复(通用错误回复)
type StandardErrorReply struct {
	Status string
}

func (r *StandardErrorReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func MakeStandardErrorReply(status string) *StandardErrorReply {
	return &StandardErrorReply{Status: status}
}

// IntReply 整数回复
type IntReply struct {
	Code int64
}

// ToBytes marshal redis.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

// MakeIntReply creates int reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// StatusReply 状态回复
type StatusReply struct {
	Status string
}

// ToBytes marshal redis.Reply
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

// MakeStatusReply creates StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
