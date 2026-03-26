package database

import (
	"godis/internal/interface/resp"
	"godis/internal/resp/reply"
	"godis/internal/util/logger"
)

type EchoDatabase struct {
}

func MakeEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	// 按照原样返回
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	logger.Info("EchoDatabase AfterClientClose")
}

func (e EchoDatabase) Close() {
	logger.Info("EchoDatabase Close")

}
