package storage

import (
	"testing"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
	"github.com/kouzant/go-short/context"
)

func TestWriteRead(t *testing.T) {
	testWriteReadBadger(t)
	testWriteReadMemory(t)	
}

func TestListAll(t *testing.T) {
	testListAllBadger(t)
	testListAllMemory(t)
}

func TestDelete(t *testing.T) {
	testDeleteBadger(t)
	testDeleteMemory(t)
}

func testWriteRead(t *testing.T, stateStore StateStore) {
	var tests = []struct{
		key string
		value string
		want StorageValue
		shouldWrite bool
		saveError error
		loadError error
	}{
		{"key0", "value0", "value0", true, nil, nil},
		{"key1", "value1", "value1", true, nil, nil},
		{"key0", "value0", "value0", true, KeyAlreadyExists{Key: "key0"}, nil},
		{"_key3", "_value3", nil, false, nil, KeyNotFound{Key: "_key3"}},
	}

	for _, test := range tests {
		item := NewStorageItem(test.key, test.value)
		if test.shouldWrite {
			saveError := stateStore.Save(item)
			if saveError != test.saveError {
				t.Errorf("stateStore.Save(%v) expected error %v - gotten %v",
					item, test.saveError, saveError)
			}
		}
		value, loadError := stateStore.Load(item.Key)
		if loadError != test.loadError {
			t.Errorf("stateStore.Load(%v) expected error %v - gotten %v",
				item, test.loadError, loadError)
		}
		if test.want != value {
			t.Errorf("stateStore.Load(%v) expected value %v - gotten %v",
				item, test.want, value)
		}
	}	
}

type A struct {
	key string
	value string
}

func testListAll(t *testing.T, stateStore StateStore) {
	numOfItems := 10
	items := make([]A, 0, numOfItems)
	for i := 0; i < numOfItems; i++ {
		items = append(items, A{fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i)})
	}

	for _, i := range items {
		item := NewStorageItem(i.key, i.value)
		err := stateStore.Save(item)
		if err != nil {
			t.Errorf("stateStore.Save(%v) failed with %v", item, err)
		}
	}

	storedItems, err := stateStore.LoadAll()
	if err != nil {
		t.Errorf("stateStore.LoadAll failed with %v", err)
	}
	for _, i := range items {
		item := NewStorageItem(i.key, i.value)
		found := false
		for _, j := range storedItems {
			if item.Key == j.Key && item.Value == j.Value {
				found = true
				break
			}
		}
		if found == false {
			t.Errorf("Item %v not found in stateStore.LoadAll() slice", item)
		}
	}
}

func testDelete(t *testing.T, stateStore StateStore) {
	item := NewStorageItem("key", "value")
	error := stateStore.Save(item)
	if error != nil {
		t.Errorf("stateStore.Save(%v) did not expect any error but gotten %v", item, error)
	}

	value, error := stateStore.Load(item.Key)
	if error != nil {
		t.Errorf("stateStore.Load(%v) did not expect any error but gotten %v", item, error)
	}
	if value != item.Value {
		t.Errorf("stateStore.Load(%v) expected value %v but gotten %v", item, item.Value, value)
	}

	value, error = stateStore.Delete(item.Key)
	if error != nil {
		t.Errorf("stateStore.Delete(%v) did not expect any error but gotten %v", item, error)
	}
	if value != item.Value {
		t.Errorf("stateStore.Delete(%v) expected to return value %v but returned %v",
			item, item.Value, value)
	}

	_, error = stateStore.Load(item.Key)
	if _, ok := error.(KeyNotFound); !ok {
		t.Errorf("Load(%v) after Delete(%v) expected %v", item, item, KeyNotFound{})
	}	
}

func testWriteReadBadger(t *testing.T) {
	stateStore, dir := createBadgerStateStore(t)
	defer os.RemoveAll(dir)
	defer stateStore.Close()
	testWriteRead(t, stateStore)
}

func testWriteReadMemory(t *testing.T) {
	stateStore := createMemoryStateStore(t)
	defer stateStore.Close()
	testWriteRead(t, stateStore)
}

func testListAllBadger(t *testing.T) {
	stateStore, dir := createBadgerStateStore(t)
	defer os.RemoveAll(dir)
	defer stateStore.Close()
	testListAll(t, stateStore)
}

func testListAllMemory(t *testing.T) {
	stateStore := createMemoryStateStore(t)
	defer stateStore.Close()
	testListAll(t, stateStore)
}

func testDeleteBadger(t *testing.T) {
	stateStore, dir := createBadgerStateStore(t)
	defer os.RemoveAll(dir)
	defer stateStore.Close()
	testDelete(t, stateStore)	
}

func testDeleteMemory(t *testing.T) {
	stateStore := createMemoryStateStore(t)
	defer stateStore.Close()
	testDelete(t, stateStore)	
}

func createBadgerStateStore(t *testing.T) (StateStore, string) {
	dir, err := ioutil.TempDir("", "test_badger_state_store")
	if err != nil {
		t.Fatal("Error creating tmp directory for Badger")
	}
	config := createConfig(dir)
	stateStore := &BadgerStateStore{Config: config}
	err = stateStore.Init()
	if err != nil {
		t.Errorf("stateStore.Init() failed with %s", err)
	}
	return stateStore, dir
}

func createConfig(dir string) *viper.Viper {
	fmt.Println("TMP dir: ", dir)
	vp := viper.New()
	vp.SetConfigType("yaml")
	vp.Set(context.StateStorePathKey, dir)
	vp.Set(context.StateStoreGCKey, "2s")
	return vp
}

func createMemoryStateStore(t *testing.T) StateStore {
	stateStore := &MemoryStateStore{}
	stateStore.Init()
	return stateStore
}
