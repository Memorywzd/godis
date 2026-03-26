package tcp

import (
	"bufio"
	"godis/internal/util/sync/atomic"
	"godis/internal/util/sync/wait"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// 定义Echo客户端
type EchoClient struct {
	Conn    net.Conn  // tcp 连接
	Sending wait.Wait // 服务端在发送数据时进入 sending 态，阻止其他goroutine关闭连接
	// TODO: 该wait实现仅仅增加了通过channel等待超时的功能，建议取消
}

// 客户端 关闭客户端连接
func (c *EchoClient) Close() {
	c.Sending.WaitWithTimeout(10 * time.Second) // TODO: 返回值未处理
	c.Conn.Close()
}

// 定义Echo服务器接口
type HandlerInterface interface {
	Handle(conn net.Conn) // 删除了ctx参数，简化接口
	Close()
}

// 定义Echo服务器
type EchoHandler struct {
	activeConn sync.Map       // 并发安全map当set，存所有正在处理的client
	isClosed   atomic.Boolean // 关闭状态标识位
	// TODO: 为何不直接用sync/atomic包中的atomic.Bool?
}

// 服务器 简单工厂模式
func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(conn net.Conn) {
	// 已关闭的 Handler 不再处理新连接
	if h.isClosed.Get() {
		conn.Close()
		return
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{}) // 存储活着的连接

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from client:", err)
			} else {
				log.Println("Connection closed by client:", conn.RemoteAddr())
				h.activeConn.Delete(client)
			}
			return
		}

		// 发送数据前置 sending 态
		client.Sending.Add(1)

		// 模拟关闭时未发送完毕数据
		log.Println("Sending fake very big message...")
		time.Sleep(5 * time.Second)

		log.Println("Received from", conn.RemoteAddr(), ":", line[:len(line)-1])
		b := []byte(line)
		_, err = conn.Write(b)
		// 发送完毕，结束 sending 态
		client.Sending.Done()
		if err != nil {
			log.Println("Error writing to client:", err)
			return
		}
	}
}

func (h *EchoHandler) Close() {
	if h.isClosed.Get() {
		return
	}
	log.Println("Close echo handler in sequence...")
	h.isClosed.Set(true)

	// 依次关闭
	h.activeConn.Range(func(key, _ any) bool {
		client := key.(*EchoClient)
		client.Close()
		return true
	})
	h.activeConn.Clear()
}
