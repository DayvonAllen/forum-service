package main

import (
	"example.com/app/events-consumer"
	"example.com/app/router"
	"fmt"
	"log"
	"os"
	"os/signal"
)

func init() {
	// create database connection instance for first time
	go event_consumer.KafkaConsumerGroup()
}

func main() {
	app := router.Setup()

	// graceful shutdown on signal interrupts
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		_ = <-c
		fmt.Println("Shutting down...")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":8085"); err != nil {
		log.Panic(err)
	}
}