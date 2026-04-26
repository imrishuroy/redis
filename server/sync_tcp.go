package server

import (
	"io"
	"log"
	"net"
	"strconv"

	"github.com/imrishuroy/redis/config"
)

func readCommand(c net.Conn) (string, error) {
	// TODO: max read in one short is 512 bytes
	// To allow input > 512 bytes then repeated read unit
	// we get EOF or designated delimiter
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:]) // blocking call: waiting for the client to send the command
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func respond(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

func RunSyncTCPServer() {
	log.Println("starting a synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	// listening to the configured host:port
	lsnr, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	// defer lsnr.Close()

	for {
		// blocking call: waiting for the new client to connect
		c, err := lsnr.Accept()
		if err != nil {
			panic(err)
		}

		// increment the number of connected clients
		con_clients++
		log.Println("client connected with address:", c.RemoteAddr(), "concurrent clients:", con_clients)

		for {
			// over the socket, continuously read the command and print it out
			cmd, err := readCommand(c)
			if err != nil {
				c.Close()
				con_clients -= 1
				log.Println("client disconnected with address:", c.RemoteAddr(), "concurrent clients:", con_clients)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}
			log.Println("command", cmd)
			if err = respond(cmd, c); err != nil {
				log.Println("err writing response:", err)
			}
		}

	}
}
