package main

import (
	"godis/internal/tcp"
)

func main() {
	tcp.ListenAndServeWithSignal(":8080", tcp.MakeEchoHandler())
}
