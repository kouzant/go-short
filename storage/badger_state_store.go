package storage

import (
	"github.com/kouzant/go-short/context"
	
	"github.com/spf13/viper"
	badger "github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
)

type BadgerStateStore struct {
	Config *viper.Viper
	db *badger.DB
}

func (s *BadgerStateStore) Init() error {
	stateStoreDir := s.Config.GetString(context.StateStoreKey)
	log.Infof("Loading state store from %s", stateStoreDir)
	db, err := badger.Open(badger.DefaultOptions(stateStoreDir))
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *BadgerStateStore) Save(item *StorageItem) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		var key []byte = []byte(string(item.Key))
		_, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				err = txn.Set([]byte(string(item.Key)), []byte(item.Value.(string)))
				return err
			}
			return err
		}
		return KeyAlreadyExists{Key: item.Key}
	})
	return err
}

func (s *BadgerStateStore) Load(key StorageKey) (StorageValue, error) {
	var valueCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(string(key)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return KeyNotFound{Key: key}
			}
			return err
		}
		valueCopy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return StorageValue(string(valueCopy)), nil
}

func (s *BadgerStateStore) Delete(key StorageKey) (StorageValue, error) {
	item, err := s.Load(key)
	if err != nil {
		if _, ok := err.(KeyNotFound); ok {
			return nil, nil
		}
		return nil, err
	}
	
	err = s.db.Update(func(txn *badger.Txn) error {
		err = txn.Delete([]byte(string(key)))
		return err
	})
	
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *BadgerStateStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
