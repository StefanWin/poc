package main

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/StefanWin/dcv/api"
	"github.com/StefanWin/dcv/scaler"
	"github.com/StefanWin/dcv/util"
)

func main() {

	if err := util.EnsureDir("./input"); err != nil {
		log.Fatal(err)
	}
	if err := util.EnsureDir("./output"); err != nil {
		log.Fatal(err)
	}

	// TODO : fix this
	e, _ := os.ReadFile(".env")
	eLines := strings.Split(string(e), "\n")
	for _, l := range eLines {
		s := strings.Split(l, "=")
		os.Setenv(s[0], s[1])
	}

	// global request channel
	requestChannel := make(chan *api.ConversionRequest, 50)

	scaler, err := scaler.NewScaler(requestChannel)
	if err != nil {
		log.Fatal(err)
	}

	if err := scaler.Initialize(); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := scaler.Start(); err != nil {
			log.Println(err)
		}
	}()

	api, err := api.NewApi(requestChannel)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := api.Listen(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("[SYSTEM]:: received shutdown signal")
	scaler.Shutdown()
	log.Println("[SYSTEM]:: shutdown scaler")
	api.Shutdown()
	log.Println("[SYSTEM]:: shutdown API")
	os.Exit(0)
}
