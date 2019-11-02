package storage

import (
	"time"

	"github.com/kouzant/go-short/context"

	badger "github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BadgerStateStore struct {
	Config *viper.Viper
	db     *badger.DB
	ticker *time.Ticker
}

func (s *BadgerStateStore) Init() error {
	stateStoreDir := s.Config.GetString(context.StateStorePathKey)
	log.Infof("Loading state store from %s", stateStoreDir)
	options := badger.DefaultOptions(stateStoreDir)
	options.Logger = log.StandardLogger()
	db, err := badger.Open(options)
	if err != nil {
		return err
	}
	s.db = db
	gcInterval, err := time.ParseDuration(s.Config.GetString(context.StateStoreGCKey))
	if err != nil {
		gcInterval = 1 * time.Hour
	}
	s.ticker = time.NewTicker(gcInterval)
	go s.startGCRoutine()

	return nil
}

func (s *BadgerStateStore) startGCRoutine() {
	for range s.ticker.C {
	again:
		err := s.db.RunValueLogGC(0.5)
		if err == nil {
			goto again
		}
	}
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

func (s *BadgerStateStore) SaveAll(items []*StorageItem) error {
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, i := range items {
		err := wb.Set([]byte(string(i.Key)), []byte(i.Value.(string)))
		return err
	}
	return wb.Flush()
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

func (s *BadgerStateStore) LoadAll() ([]*StorageItem, error) {
	storedItems := make([]*StorageItem, 0, 100)

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			keyCopy := item.KeyCopy(nil)
			valueCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			storedItems = append(storedItems, NewStorageItem(string(keyCopy), string(valueCopy)))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return storedItems, nil
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
	s.ticker.Stop()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
