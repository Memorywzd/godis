package server

import (
	"godis/internal/database"
	idatabse "godis/internal/interface/database"
	"godis/internal/resp"
	"godis/internal/resp/reply"
	"godis/internal/util/logger"
	"godis/internal/util/sync/atomic"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler implements tcp.Handler and serves as a redis handler
type RespHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         idatabse.Database
	closing    atomic.Boolean // refusing new client and new request
}

// MakeRespHandler creates a RespHandler instance
func MakeRespHandler() *RespHandler {
	var db idatabse.Database
	db = database.MakeEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

// Handle receives and executes redis commands
func (h *RespHandler) Handle(conn net.Conn) {
	if h.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := MakeConnection(conn)
	h.activeConn.Store(client, 1)

	ch := resp.ParseStream(conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeStandardErrorReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

func (h *RespHandler) closeClient(client *Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Close stops handler
func (h *RespHandler) Close() {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
}
