package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	webhook "github.com/jimtang2/grok-webhook"
	"github.com/spf13/viper"
)

var handler = &webhook.Handler{
	Projects: map[string]string{},
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("config reloaded:", e.Name)
		handler.Projects = viper.GetStringMapString("projects")
	})
	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	handler.Projects = viper.GetStringMapString("projects")
}

func main() {
	log.Println("webhook started; listening on :8080")
	http.ListenAndServe(":8080", handler)
}
