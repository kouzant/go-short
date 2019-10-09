package storage

import (
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
)

type MemoryStateStore struct {
	Config *viper.Viper
	db map[StorageKey]StorageValue
}

func (s *MemoryStateStore) Init() error {
	log.Info("Initializing memory state store")
	s.db = make(map[StorageKey]StorageValue)
	return nil
}

func (s *MemoryStateStore) Save(item *StorageItem) error {
	if _, ok := s.db[item.Key]; ok {
		return KeyAlreadyExists{Key: item.Key}
	}
	s.db[item.Key] = item.Value
	return nil
}

func (s *MemoryStateStore) Load(key StorageKey) (StorageValue, error) {
	if value, ok := s.db[key]; ok {
		return value, nil
	}
	return nil, KeyNotFound{Key: key}
}

func (s *MemoryStateStore) Delete(key StorageKey) (StorageValue, error) {
	if value, ok := s.db[key]; ok {
		delete(s.db, key)
		return value, nil
	}
	return nil, nil
}

func (s *MemoryStateStore) Close() error {
	s.db = make(map[StorageKey]StorageValue)
	return nil
}
