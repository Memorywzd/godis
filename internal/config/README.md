## Config 加载

Redis 中的配置文件是 `redis.conf`，用于存储 Redis 服务器的配置信息。

- `bind`：指定 Redis 服务器监听的 IP 地址和端口号。
- `port`：指定 Redis 服务器监听的端口号。
- `appendonly`：指定是否启用 AOF 持久化。
- `appendfilename`：指定 AOF 持久化文件的名称。
- `maxclients`：指定 Redis 服务器的最大客户端连接数。
- `databases`：指定 Redis 服务器的数据库数量。
- `requirepass`：指定 Redis 服务器的密码。
- `peers`：指定 Redis 服务器的对等节点。
- `self`：指定 Redis 服务器的自身节点。
