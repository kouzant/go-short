package main

import (
	"github.com/kouzant/go-short/context"
	"github.com/kouzant/go-short/logger"
	"github.com/kouzant/go-short/storage"	
	log "github.com/sirupsen/logrus"	
)

func main() {
	conf := context.ReadConfig()
	logger.Init(conf)
	log.Info("Starting go-short")

	//stateStore := &storage.StateStore{Config: conf}
	stateStore := &storage.MemoryStateStore{Config: conf}
	stateStore.Init()
	write(stateStore, "antonis", "kouzoupis")
	load(stateStore, "antoniss")
}
