这里主要创建了 `Debug`、`Info`、`Warn`、`Error`、`Fatal` 五个函数，用于打印不同级别的日志。

每个日志的处理函数中，都使用了 `mu.Lock()` 和 `mu.Unlock()` 来保证线程安全。logger 是一个全局变量，如果多个 goroutine 同时调用这些函数，可能会导致日志输出的顺序不一致。
使用 `mu.Lock()` 和 `mu.Unlock()` 可以保证每个 goroutine 在调用日志函数时，不会被其他 goroutine 打断。

logger 需要调用 io 读写本地文件，因此需要创建一个 `mustOpen` 函数，用于打开文件。