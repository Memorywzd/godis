package database

import (
	"godis/internal/datastruct/dict"
	"godis/internal/interface/database"
	"godis/internal/interface/datastruct"
	"godis/internal/interface/resp"
	"godis/internal/resp/reply"
	"godis/internal/util/logger"
	"strconv"
	"strings"
)

type RedisDatabase struct {
	index int
	data  datastruct.Dict
}

// MakeRedisDatabase creates a new RedisDatabase instance
func MakeRedisDatabase() *RedisDatabase {
	return &RedisDatabase{
		index: 0,
		data:  dict.MakeSyncDict(),
	}
}

// ExecFunc is a function type that takes a RedisDatabase instance and a slice of byte slices as arguments and returns a resp.Reply
// All redis commands like PING, SET, GET, etc. are implemented as functions of this type
type ExecFunc func(db *RedisDatabase, args [][]byte) resp.Reply

// CmdLine is a type alias for a slice of byte slices
// It is used to represent the command line arguments passed to the ExecFunc
type CmdLine = [][]byte

// Exec executes a command on the RedisDatabase instance
// It takes a connection and a command line as arguments
// It returns a resp.Reply which is the response to the command
func (db *RedisDatabase) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	// The first element of cmdLine is the command name, like "PING", "SET", etc.
	// Convert it to lowercase to ensure case-insensitivity
	cmdName := strings.ToLower(string(cmdLine[0]))
	// Get the command from the command table using the command name
	// If the command is not found, return an error reply
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}
	// Validate the number of arguments passed to the command
	if !ValidateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	// Execute the command and return the response
	return cmd.exec(db, cmdLine[1:])
}

// ValidateArity checks if the number of arguments passed to a command is valid
func ValidateArity(arity int, args [][]byte) bool {
	// Check if the number of arguments is less than the required arity
	if arity >= 0 {
		return len(args) == arity
	} else {
		// If the arity is negative, it means the command takes a variable number of arguments
		// Check if the number of arguments is within the valid range
		return len(args) >= -arity
	}
}

// GetEntity returns DataEntity bind to the given key
func (db *RedisDatabase) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// PutEntity stores the given DataEntity in the database
func (db *RedisDatabase) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists edit the given DataEntity in the database
func (db *RedisDatabase) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent stores the given DataEntity in the database if it doesn't already exist
func (db *RedisDatabase) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove deletes the DataEntity associated with the given key from the database
func (db *RedisDatabase) Remove(key string) int {
	return db.data.Remove(key)
}

// Removes deletes the DataEntity associated with the given keys from the database
func (db *RedisDatabase) Removes(keys ...string) int {
	deleted := 0
	for _, key := range keys {
		// Use Remove's return value directly to avoid race condition between Get and Remove
		result := db.data.Remove(key)
		if result > 0 {
			deleted++
		}
	}
	return deleted
}

// Flush clears the database by removing all DataEntity objects
func (db *RedisDatabase) Flush() {
	db.data.Clear()
}

type Database struct {
	dbSet *RedisDatabase
}

// MakeDatabase creates a new Database instance
func MakeDatabase() *Database {
	_database := &Database{}
	_database.dbSet = MakeRedisDatabase()
	return _database
}

// execSelect sets the current database for the client connection.
// select x 由于仅仅实现了一个db，select多少都会返回1
func execSelect(c resp.Connection, database *Database, args [][]byte) resp.Reply {
	_, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid RedisDatabase index")
	}
	/*
		if dbIndex < 0 || dbIndex >= len(database.dbSet) {
			return reply.MakeStandardErrorReply("ERR RedisDatabase index out of range")
		}
		c.SelectDB(dbIndex)
	*/
	c.SelectDB(1) // 仅实现一个db，select多少都会返回1
	return reply.MakeIntReply(int64(1))
}

// Exec executes a command on the database
func (d *Database) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Database Exec panic:" + err.(error).Error())
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, d, args[1:])
	}
	// Get the current database index from the client connection
	// db := d.dbSet[client.GetDBIndex()]
	return d.dbSet.Exec(client, args)
}

func (d *Database) AfterClientClose(c resp.Connection) {

}

func (d *Database) Close() {

}
