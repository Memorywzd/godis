package tcp

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// 监听并交给handler处理，收到closeChan的通知后关闭
func ListenAndServeWithSignal(addr string, handler HandlerInterface) {
	closeChan := make(chan struct{}) // TODO: 与 make(chan bool, 1) 的区别
	signalChan := make(chan os.Signal)
	/*
		SIGINT：2 中断信号（通常是 Ctrl+C）
		SIGHUP：1 终端断开
		SIGQUIT：3 退出信号
		SIGTERM：15 终止信号
	*/
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// 信号处理
	go func() {
		s := <-signalChan
		log.Println("Received signal", s)
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen on address", addr, ":", err)
	}
	log.Println("Started listening on", addr)

	// 关闭处理，原先逻辑为 defer listener.Close()
	go func() {
		<-closeChan
		log.Println("Received closeChan, shutting down gracefully...")
		_ = listener.Close() // 停止tcp监听，listener.Accept()会立即返回 io.EOF
		handler.Close()
	}()
	defer func() {
		_ = listener.Close()
		handler.Close() // TODO: 重复关闭
	}()

	// 创建空的contex TODO: 由于具体实现并未使用上下文，删掉
	// ctx := context.Background()
	var serveDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			// 通常是由于listener被关闭或发生网络错误引起的
			log.Println("Error accepting connection:", err)
			break
		}
		log.Println("Accepted new connection from", conn.RemoteAddr())
		serveDone.Add(1)
		go func() {
			defer serveDone.Done()
			handler.Handle(conn)
		}()
	}
	serveDone.Wait()
}
