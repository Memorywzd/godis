package main

import (
	"godis/internal/tcp"
)

func main() {
	tcp.ListenAndEchoWithSignal(":8080", tcp.MakeEchoHandler())
}
