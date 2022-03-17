package main

import (
	"github.com/samuelmahr/appt-scheduling/internal/application"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	"log"
)

func main() {
	c, err := configuration.Configure()
	if err != nil {
		log.Fatal(err)
	}

	app := application.NewAPIApplication(c)
	app.Run()
}
