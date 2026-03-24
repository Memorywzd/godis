package main

import (
	"godis/internal/tcp"
)

func main() {
	tcp.ListenAndEcho("localhost:8080")
}
