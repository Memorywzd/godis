package database

import (
	"godis/internal/interface/resp"
	"godis/internal/resp/reply"
)

// Register the PING command to the command table
func init() {
	// Register the PING command with the command table
	RegisterCommand("ping", Ping, 1)
}

func Ping(db *RedisDatabase, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}
