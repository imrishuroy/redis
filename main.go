package main

import (
	"flag"
	"log"

	"github.com/imrishuroy/redis/config"
	"github.com/imrishuroy/redis/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "the host to listen on")
	flag.IntVar(&config.Port, "port", 6379, "the port to listen on")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Printf("starting the server on %s:%d\n", config.Host, config.Port)
	server.RunSyncTCPServer()
}


