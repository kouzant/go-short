package storage

import (
	"testing"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

func TestWriteReadBadger(t *testing.T) {
	dir, err := ioutil.TempDir("", "test_badger_state_store")
	if err != nil {
		t.Error("Error creating tmp directory for Badger")
	}
	config := createConfig(dir)
	stateStore := &BadgerStateStore{Config: config}
	defer func(){
		os.RemoveAll(dir)
		stateStore.Close()
	}()
	err = stateStore.Init()
	if err != nil {
		t.Errorf("stateStore.Init() failed with %s", err)
	}
	defer stateStore.Close()

	item := &StorageItem{StorageKey("name"), StorageValue("antonis")}
	err = stateStore.Save(item)
	if err != nil {
		t.Errorf("save error")
	}

	val, err := stateStore.Load(item.Key)
	if err != nil {
		t.Errorf("load error")
	}
	fmt.Println("Value ", val)
}

func createConfig(dir string) *viper.Viper {
	fmt.Println("TMP dir: ", dir)
	vp := viper.New()
	vp.SetConfigType("yaml")
	vp.Set("go-short.state-store", dir)
	return vp
}
