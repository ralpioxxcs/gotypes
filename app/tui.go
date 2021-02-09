package app

import (
	"log"
	"os"
)

var Logger log.Logger

func Run() {
	a := NewApp()

	f, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		Logger.Fatal(err)
	}
	defer f.Close()
	Logger.SetOutput(f)
	Logger.Println("-----------------------------")

	// Run tui main
	if err := a.Run(); err != nil {
		panic(err)
	}
}
