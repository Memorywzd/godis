package resp

import (
	"bufio"
	"errors"
	"godis/internal/interface/resp"
	"godis/internal/resp/reply"
	"godis/internal/util/logger"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type readState struct {
	readingMultiLine  bool     // 是否正在读取多行数据
	expectedArgsCount int      // 期望的参数数量
	msgType           byte     // 消息类型
	args              [][]byte // 参数
	bulkLen           int64    // Bulk 回复的长度
}

func (r *readState) isDone() bool {
	return r.expectedArgsCount > 0 && len(r.args) == r.expectedArgsCount
}

type Payload struct {
	Data resp.Reply // 客户端发给服务端的和服务端发给客户端的数据使用的是一个结构，因此也能用 Reply 接口
	Err  error
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// 多行读取的
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // err reply
		result = reply.MakeStandardErrorReply(str[1:])
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// 赋值，表示下一行中的 Bulk 回复的长度
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}

func readRespLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var line []byte
	var err error
	if state.bulkLen == 0 {
		line, err = bufReader.ReadBytes('\n')
		if err != nil {
			// TODO: 别扭的错误处理
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' {
			// 不符合RESP协议的行：\n已经由ReadBytes处理了，所以我们需要检查倒数第二个字符是否是\r
			return nil, false, errors.New("invalid line terminator")
		}
	} else {
		line = make([]byte, state.bulkLen+2)  // 包括\r\n
		_, err = io.ReadFull(bufReader, line) // 读len(line)字节到line中 TODO: 判断读取的字节数与state.bulkLen的关系
		if err != nil {
			// 同样别扭
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			return nil, false, errors.New("invalid line terminator")
		}
		// state.bulkLen -= int64(len(line) - 2)
		state.bulkLen = 0 // 为何置空？
	}
	return line, false, nil
}

func parse(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			// 打印调用栈信息
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader) // 读取缓冲区
	var state readState                  // 解析器的状态
	var err error
	var msg []byte

	// 读取数据
	for {
		var ioErr bool // 是否是 IO 错误
		msg, ioErr, err = readRespLine(bufReader, &state)

		if err != nil {
			// 如果是 IO 错误，关闭通道，退出循环
			if ioErr {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			ch <- &Payload{Err: err}
			state = readState{} // 重置状态
			continue            // 继续循环，读取下一行
		}

		if !state.readingMultiLine {
			// 多条批量回复
			switch msg[0] {
			case '*':
				// 解析头部，获取期望的参数数量
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
				// 需要的参数数量为 0，直接返回
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
			case '$':
				// Bulk 回复
				err = parseBulkHeader(msg, &state) // 解析 Bulk 回复的头部，获取 Bulk 回复的长度
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
				if state.bulkLen == -1 {
					// Bulk 回复的长度为 0，直接返回
					ch <- &Payload{Data: &reply.NullBulkReply{}}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
			default:
				// 单行回复
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{} // 本条消息已结束，重置状态
				continue            // 继续循环，读取下一行
			}
		} else {
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// 如果满足 isDone 条件，表示解析完成
			// 创建一个回复，发送给客户端
			if state.isDone() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}

	// ...
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse(reader, ch)
	return ch
}
