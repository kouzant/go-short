package storage

import (
	"fmt"
)

type StorageKey string
type StorageValue interface{}

type StorageItem struct {
	Key   StorageKey
	Value StorageValue
}

type Pair struct {
	Left  interface{}
	Right interface{}
}

func NewStorageItem(key, value string) *StorageItem {
	return &StorageItem{StorageKey(key), StorageValue(value)}
}

type KeyAlreadyExists struct {
	Key StorageKey
}

func (e KeyAlreadyExists) Error() string {
	return fmt.Sprintf("Key %s already exists", e.Key)
}

type KeyNotFound struct {
	Key StorageKey
}

func (e KeyNotFound) Error() string {
	return fmt.Sprintf("Key %s does not exist", e.Key)
}

type StateStore interface {
	Init() error
	Save(item *StorageItem) error
	SaveAll(items []*StorageItem) error
	Load(key StorageKey) (StorageValue, error)
	LoadAll() ([]*StorageItem, error)
	Delete(key StorageKey) (StorageValue, error)
	Close() error
}
