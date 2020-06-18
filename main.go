package main

import (
	"github.com/dombo/srebot/pkg/bot"
	"log"
)

func main() {
	b, err := bot.NewBot(bot.GetConf())
	if err != nil {
		log.Fatal(err)
	}

	err = b.Run()
	if err != nil {
		b.Logger.Fatal(err.Error())
	}
}