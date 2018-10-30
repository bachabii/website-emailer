package main

import (
	"log"
	"fmt"
	"encoding/json"
	"os"

	"github.com/streadway/amqp"
	"net/smtp"
)

type Message struct {
	Email string
	Subject string
	Content string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func send(email string, subject string, content string) {
	log.Printf("hi there emailer")
	var from, pass, to string
	
	if ( os.Getenv("GO_ENV") == "production" ) {
		from = os.Getenv("EMAIL_FROM")
		pass = os.Getenv("EMAIL_PWD")
		to = os.Getenv("EMAIL_TO")
	} else {
		from = "ibac7889@gmail.com"
		pass = "Juillet7889!"
		to = "ibac7889@gmail.com"
	}

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Website Msg ("+ email +"): " + subject + "\n\n" +
		content

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}	
}

func main() {
	var amqp_url string
	if ( os.Getenv("GO_ENV") == "production" ) {
		amqp_url = os.Getenv("AMQP_URL")
	} else {
		amqp_url = "amqp://guest:guest@localhost:5672/"
	}
	
	conn, err := amqp.Dial(amqp_url)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello_test", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	

	go func() {
		for d := range msgs {
			var m Message
			err := json.Unmarshal(d.Body, &m)

			if err == nil {
				fmt.Printf("%+v\n", err)
				fmt.Printf("%+v\n", m.Email)
				log.Printf("Received a message: %s", d.Body)

				send(m.Email, m.Subject, m.Content)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}