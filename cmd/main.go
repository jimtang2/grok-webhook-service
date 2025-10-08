package main

import (
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
		if len(handler.Projects) == 0 {
			log.Println("no configured project")
		}
	})
	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("no config file: %w", err)
	}
	handler.Projects = viper.GetStringMapString("projects")
	if len(handler.Projects) == 0 {
		log.Println("no configured project")
	}
}

func main() {
	log.Println("webhook started; listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
