package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"listener_cache_writethrough/src/config"
	"listener_cache_writethrough/src/listener"
)

func gracefulShutdown(lstr listener.Listener, c chan os.Signal, cancel context.CancelFunc) {
	<-c
	fmt.Println("\ngraceful shutdown initiated")

	cancel()
	lstr.CloseDB()
	lstr.CloseChan()

	os.Exit(0)
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.New(config.ENV)
	if err != nil {
		panic(err)
	}

	lstr := listener.New(cfg, ctx)

	go gracefulShutdown(lstr, c, cancel)

	lstr.Start(90)

}
