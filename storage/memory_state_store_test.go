package storage

import (
	"testing"
)

func TestWriteReadMemory(t *testing.T) {
	stateStore := &MemoryStateStore{}
	stateStore.Init()
	defer stateStore.Close()

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

func TestDeleteMemory(t *testing.T) {
	stateStore := &MemoryStateStore{}
	stateStore.Init()
	defer stateStore.Close()

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
