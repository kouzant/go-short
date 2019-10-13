package logger

import (
	"fmt"
	"os"

	"github.com/kouzant/go-short/context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init(config *viper.Viper) {
	log_level := config.GetString(context.LogLevelKey)
	lvl, err := logrus.ParseLevel(log_level)
	if err != nil {
		panic(fmt.Errorf("Fatal error %s\n", err))
	}
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(lvl)
}
