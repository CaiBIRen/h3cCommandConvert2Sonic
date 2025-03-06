package configuration

import (
	"fmt"
	h3cmodel "sonic-unis-framework/model/h3c"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Configstru struct {
	Configmux   sync.RWMutex
	Company     string                  `json:"Company"`
	Role        string                  `json:"Role"`
	Serverlldps []h3cmodel.LLDPNeighbor `json:"Serverlldps"`
	Vfws        []Vfwinfo               `json:"Vfws"`
}

type Vfwinfo struct {
	Name     string `json:"Name"`
	IP       string `json:"IP"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

var ServiceConfiguration = Configstru{Company: "H3C", Role: "Unconfig", Serverlldps: make([]h3cmodel.LLDPNeighbor, 0), Vfws: make([]Vfwinfo, 0)}

func ConfigViper() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/sonic-unis-framework/")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		fmt.Printf("Failed to read config %v\n", err) // Handle errors reading the config file
		panic("Failed to read config")
	}

	if err := viper.Unmarshal(&ServiceConfiguration); err != nil {
		panic("viper unmarshal failed")
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		ServiceConfiguration.Configmux.Lock()
		defer ServiceConfiguration.Configmux.Unlock()
		if err := viper.Unmarshal(&ServiceConfiguration); err != nil {
			panic("viper unmarshal failed")
		}
	})
}

func ViperMutexWriteConfig(name string, value string) {
	ServiceConfiguration.Configmux.Lock()
	defer ServiceConfiguration.Configmux.Unlock()
	viper.Set(name, value)
	viper.WriteConfig()
}

func ViperSetKeyValue2Cache(name string, value string) {
	ServiceConfiguration.Configmux.Lock()
	defer ServiceConfiguration.Configmux.Unlock()
	viper.Set(name, value)
}

func ViperGetValueFromCache(name string) string {
	ServiceConfiguration.Configmux.Lock()
	defer ServiceConfiguration.Configmux.Unlock()
	value := viper.GetString(name)
	return value
}

// func ViperUnsetKeyValue2Cache(name string) interface{} {
// 	ServiceConfiguration.Configmux.Lock()
// 	defer ServiceConfiguration.Configmux.Unlock()
// }

