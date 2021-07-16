package main

import (
	"github.com/meekyphotos/experive-cli/core/initializers"
	"log"
	"os"
)

func main() {
	app := initializers.InitApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
