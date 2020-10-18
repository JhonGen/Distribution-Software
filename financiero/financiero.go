package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	mensajes, err := ch.Consume(
		"Finanzas",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	forever := make(chan bool)
	go func() {
		for d := range mensajes {
			fmt.Printf("mensaje recibido: %s\n", d.Body)
		}
	}()

	fmt.Printf("Succesfully connected to pichula\n")
	fmt.Printf("esperando mas pichula\n")
	<-forever
}
