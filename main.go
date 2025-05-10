package main

import (
	"fmt"
	"mall/handler"
	"mall/storage"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func main() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	e := gin.Default()

	storage.Init("db")
	defer storage.Close()

	handler.NewHandler(e, cfg.Mini.AppID, cfg.Mini.Secret, cfg.Jwt.Secret)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go e.Run(":8083")

	<-ch

}
