package main

import (
	"context"

	"listener_cache_writethrough/src/config"
	"listener_cache_writethrough/src/listener"
)

func main() {

	cfg, err := config.New(config.ENV)
	if err != nil {
		panic(err)
	}

	lstr := listener.New(cfg, context.Background())
	lstr.Start()

}
