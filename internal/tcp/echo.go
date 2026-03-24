package tcp

import (
	"bufio"
	"io"
	"log"
	"net"
)

func HandleEcho(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞，直到遇到指定的分隔符
		// 遇到分隔符后会返回上次遇到分隔符或连接建立后收到的所有数据，包括分隔符本身
		// 若在遇到分隔符之前就发生了错误（如连接关闭），则会返回已读取的数据和错误
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from client:", err.Error())
			} else {
				log.Println("Connection closed by client:", conn.RemoteAddr())
			}
			return
		}
		log.Println("Received from", conn.RemoteAddr(), ":", line)
		// 字符串转换为字节切片（字节流）
		b := []byte(line)
		// 原样返回给客户端，写回
		_, err = conn.Write(b)
		if err != nil {
			log.Println("Error writing to client:", err)
			return
		}
	}
}

func ListenAndEcho(addr string) {
	// 绑定监听地址
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen on address", addr, ":", err)
	}
	defer listener.Close()
	log.Println("Started listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			// 通常是由于listener被关闭或发生网络错误引起的
			log.Println("Error accepting connection:", err)
		}
		log.Println("Accepted new connection from", conn.RemoteAddr())
		go HandleEcho(conn)
	}
}

/*
拆包与粘包问题

我们常说的 TCP 服务器并非「实现 TCP 协议的服务器」而是「基于TCP协议的应用层服务器」。
TCP 是面向字节流的协议，而应用层协议大多是面向消息的，比如 HTTP 协议的请求/响应，Redis 协议的指令/回复都是以消息为单位进行通信的。

作为应用层服务器我们有责任从 TCP 提供的字节流中正确地解析出应用层消息，在这一步骤中我们会遇到「拆包/粘包」问题。

socket 允许我们通过 read 函数读取新收到的一段数据（当然这段数据并不对应一个 TCP 包）。
在上文的 Echo 服务器示例中我们用\n表示消息结束，从 read 函数读取的数据可能存在下列几种情况:

收到两段数据: "abc", "def\n" 它们属于一条消息 "abcdef\n" 这是被拆包的情况
收到一段数据: "abc\ndef\n" 它们属于两条消息 "abc\n", "def\n" 这是粘包的情况
应用层协议通常采用下列几种思路之一来定义消息，以保证完整地进行读取:

- 定长消息；
- 在消息尾部添加特殊分隔符，如示例中的Echo协议和FTP控制协议。
  bufio 标准库会缓存收到的数据直到遇到分隔符才会返回，它可以帮助我们正确地分割字节流。
- 将消息分为 header 和 body, 并在 header 中提供 body 总长度，这种分包方式被称为 LTV(length，type，value) 包
  这是应用最广泛的策略，如HTTP协议。当从 header 中获得 body 长度后, io.ReadFull 函数会读取指定长度字节流，从而解析应用层消息。

在没有具体应用层协议的情况下，我们很难详细地讨论拆包与粘包问题。
Redis 序列化协议(RESP)结合应用分隔符和 LTV 包两种策略来定义消息格式，既保证了消息的完整性又提高了协议的灵活性。
*/
