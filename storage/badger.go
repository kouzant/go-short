package storage

import (
	"github.com/kouzant/go-short/context"
	
	"github.com/spf13/viper"
	//badger "github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

type BadgerStateStore struct {
	Config *viper.Viper
}

func (s *BadgerStateStore) Init() error {
	stateStore := s.Config.GetString(context.StateStoreKey)
	log.Infof("Loading state store from %s", stateStore)
	return nil
}

