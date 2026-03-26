package main

import (
	"fmt"
	"godis/internal/config"
	"godis/internal/server"
	"godis/internal/tcp"
	"godis/internal/util/logger"
	"os"
)

const configFile string = "configs/redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	tcp.ListenAndServeWithSignal(
		config.Properties.Bind+":"+fmt.Sprint(config.Properties.Port),
		server.MakeRespHandler())

}
