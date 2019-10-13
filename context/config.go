package context

import (
	"fmt"
	"os/user"

	"github.com/spf13/viper"
)

const (
	AppName = "go-short"

	configRoot = AppName + "."

	LogLevelKey = configRoot + "log-level"

	stateStore        = configRoot + "state-store."
	StateStorePathKey = stateStore + "path"
	StateStoreGCKey   = stateStore + "gc-interval"

	web          = configRoot + "webserver."
	WebListenKey = web + "listen"
	WebPortKey   = web + "port"

	CLI_USER_AGENT = "go-short-cli"
)

func ReadConfig() *viper.Viper {
	viper := viper.GetViper()
	viper.SetConfigType("yaml")
	viper.SetConfigName("go-short")
	viper.AddConfigPath("$HOME/.go-short")

	viper.SetDefault(LogLevelKey, "info")
	viper.SetDefault(StateStorePathKey, "~/.go-short/state-store")
	viper.SetDefault(StateStoreGCKey, "1h")
	viper.SetDefault(WebListenKey, "localhost")
	viper.SetDefault(WebPortKey, "80")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error: %s\n", err))
	}
	return viper
}

func getUserHome() string {
	user, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return user.HomeDir
}
