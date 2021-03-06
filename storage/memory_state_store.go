package storage

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type MemoryStateStore struct {
	Config *viper.Viper
	db     map[StorageKey]StorageValue
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

func (s *MemoryStateStore) SaveAll(items []*StorageItem) error {
	for _, i := range items {
		s.db[i.Key] = i.Value
	}

	return nil
}

func (s *MemoryStateStore) Load(key StorageKey) (StorageValue, error) {
	if value, ok := s.db[key]; ok {
		return value, nil
	}
	return nil, KeyNotFound{Key: key}
}

func (s *MemoryStateStore) LoadAll() ([]*StorageItem, error) {
	storedItems := make([]*StorageItem, 0, len(s.db))
	for key, value := range s.db {
		storedItems = append(storedItems, &StorageItem{key, value})
	}
	return storedItems, nil
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
