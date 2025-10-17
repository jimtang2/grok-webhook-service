// webhook$$grok-webhook;main.go;grok$$
package main

import (
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	webhook "github.com/jimtang2/grok-webhook"
	"github.com/spf13/viper"
)

var (
	handler = &webhook.Handler{
		Projects: map[string]string{},
		Messages: make(chan *webhook.Message),
	}
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("port", ":8080")
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
	go func() {
		err := http.ListenAndServe(viper.GetString("port"), handler)
		log.Fatal(err)
	}()
	model := NewModel()
	go model.run()

	for {
		select {
		case msg := <-handler.Messages:
			go func() {
				model.Update(msg)
			}()
		}
	}
}
