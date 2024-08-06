package examples

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var GlobalConfig *viper.Viper

func init() {
	fmt.Printf("Loading configuration logics...\n")
	GlobalConfig = initConfig()
	go dynamicConfig()
}

func initConfig() *viper.Viper {
	GlobalConfig := viper.New()
	GlobalConfig.SetConfigName("test")
	GlobalConfig.AddConfigPath("conf/")
	GlobalConfig.SetConfigType("json")
	err := GlobalConfig.ReadInConfig()
	if err != nil {
		fmt.Printf("Failed to get the configuration.")
	}
	return GlobalConfig
}

func dynamicConfig() {
	GlobalConfig.WatchConfig()
	GlobalConfig.OnConfigChange(func(event fsnotify.Event) {

		fmt.Printf("Detect config change: %s \n", event.String())
	})
}
