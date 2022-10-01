package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	name    = "EchoTool"
	version = "1.0.0"
)

func main() {
	cli := cli{}
	config, err := cli.execute()
	if err != nil {
		fmt.Printf("%s: error: %s\n", cli.binName, err)
		fmt.Printf("Type %s -help to see a list of all options.", cli.binName)
		os.Exit(1)
	}

	if config.serverMode {
		server := server{
			config: config,
			sigint: sigint(),
		}

		server.run()
	} else {
		client := client{
			config: config,
			sigint: sigint(),
		}

		client.run()
	}
}

func sigint() *chan bool {
	sigint := make(chan os.Signal, 1)
	close := make(chan bool, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigint
		close <- true
	}()

	return &close
}
