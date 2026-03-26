LF (Line Feed, 换行)：ASCII 码值为 10 (0x0A)，在 Unix 和类 Unix 系统 (macOS, Linux) 中用作换行符。
它表示将光标移动到下一行开头。

CRLF (Carriage Return Line Feed, 回车换行)：由两个字符组成。
CR (Carriage Return, 回车) ASCII 码值为 13 (0x0D) 和 LF (Line Feed, 换行) ASCII 码值为 10 (0x0A)。
在 Windows 系统中用作换行符。它表示将光标先移动到当前行开头 (CR)，然后再移动到下一行开头 (LF)。

consts主要需要实现的是：

- `PongReply`：在客户端发送 `PING` 命令时的回复是固定的 `PONG`。
- `OKReply`：在客户端发送 `SET` 命令时的回复是固定的 `OK`。
- `NullBulkReply`：空的 Bulk 回复，Bulk 是多行字符串，`-1` 表示 `nil` 值。比如 `GET` 命令，如果 key 不存在，就会返回 `nil`。
- `EmptyBulkReply`：空的 Bulk 回复，`0` 表示空字符串。比如 `SET` 命令，如果 key 不存在，就会返回空字符串。
- `EmptyMultiBulkReply`：空的 MultiBulk 回复，`0` 表示空数组。比如 `LRANGE` 命令，如果 key 不存在，就会返回空数组。
- `NoReply`：无回复。

> 在 Redis 中，Bulk 是多行字符串，MultiBulk 是数组。

error回复基本也是固定的，只是会多加一些用户输入的命令参数等信息来帮助用户定位错误。

error主要需要实现的是：

- `ArgNumErrReply`：参数数量错误回复。当用户输入的命令参数数量不正确时，可以返回这个回复。
- `UnknownReply`：未知错误回复。当我们不知道错误是什么时，可以返回这个回复。
- `SyntaxErrReply`：语法错误回复。当用户输入的命令有语法错误时，可以返回这个回复。
- `WrongTypeErrReply`：类型错误回复。当用户对一个错误的数据类型执行操作时，可以返回这个回复。
- `ProtocolErrReply`：协议错误回复。当用户输入的命令有协议错误时，例如对于数组需要 `*` 开头，对于字符串需要 `$` 开头，而用户没有遵守这个规则时，可以返回这个回复。

reply主要需要实现的是：

- `BulkReply`：用于处理多行字符串回复，结构体中存储的是想要返回的字符串，然后实现 Reply 接口的 ToBytes 方法，将字符串转换为符合 RESP 协议的字节切片。
- `MultiBulkReply`：用于处理数组回复，遍历数组，然后将数组中的每个字符串转换为 RESP 协议的字节切片。然后将这些字节切片拼接起来，返回。
- `StandardErrorReply`：标准错误回复。当我们需要返回一个错误信息时，可以使用这个回复。
- `IntReply`：整数回复。当我们需要返回一个整数时，可以使用这个回复。
- `StatusReply`：状态回复。当我们需要返回一个状态时，可以使用这个回复。
