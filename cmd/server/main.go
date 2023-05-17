package main

import (
	"context"
	"flag"

	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/api"
)

func main() {
	var (
		configPath string
		serverID   string
	)
	// flag.StringVar(&configPath, "config", "./etc/config.yaml", "path to the server config file")
	flag.StringVar(&configPath, "config", "../../etc/server.yaml", "path to the server config file")
	flag.StringVar(&serverID, "sid", "server", "server id")
	flag.Parse()

	logging.GlobalLogger.Infof(context.Background(), "start server with config file %s and server id %s", configPath, serverID)
	err := api.Entry(serverID, configPath)
	if err != nil {
		panic(err)
	}
}
