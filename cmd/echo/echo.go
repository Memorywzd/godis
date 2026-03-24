package main

import (
	"godis/internal/tcp"
)

func main() {
	tcp.ListenAndEcho(":8080")
}
