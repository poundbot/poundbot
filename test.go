package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/go-acme/lego/log"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("poundbot")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/poundbot/")
	viper.AddConfigPath("$HOME/.poundbot/")

	fmt.Println(viper.ConfigFileUsed())
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		log.Warnf("Error reading config file: %s,%s", reflect.TypeOf(err), err)
		os.Exit(1)
	}
	fmt.Println(viper.ConfigFileUsed())
}
