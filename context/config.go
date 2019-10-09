package context

import (
	"fmt"
	
	"github.com/spf13/viper"
)

const (
	AppName = "go-short"

	configRoot = AppName + "."

	LogLevelKey = configRoot + "log-level"
	StateStoreKey = configRoot + "state-store"
	
	web = configRoot + "webserver."
	WebListenKey = web + "listen"
	WebPortKey = web + "port"
)

func ReadConfig() *viper.Viper {
	viper := viper.GetViper()
	viper.SetConfigType("yaml")
	viper.SetConfigName("go-short")
	viper.AddConfigPath("$HOME/.go-short")

	viper.SetDefault(LogLevelKey, "info")
	viper.SetDefault(StateStoreKey, "~/.go-short/state-store")
	viper.SetDefault(WebListenKey, "localhost")
	viper.SetDefault(WebPortKey, "80")
	
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error: %s\n", err))
	}
	return viper
}
