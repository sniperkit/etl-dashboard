package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"time"
	"net/http"
	"encoding/json"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	bytes := make([]byte, l)

	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}

	return string(bytes)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var rabbitUser = flag.String("user", "guest", "RabbitMQ user name")
	var rabbitPassword = flag.String("password", "guest", "RabbitMQ password")
	var rabbitHost = flag.String("host", "localhost", "RabbitMQ host")
	var rabbitPort = flag.String("port", "5672", "RabbitMQ port")

	flag.Parse()

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s", *rabbitUser, *rabbitPassword, *rabbitHost, *rabbitPort))

	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to retrieve connection to rabbitmq server at %v:%v using provided credentials", rabbitHost, rabbitPort))
		panic(err)

	}
	defer conn.Close()
	exchangeName := "etl_exchange"
	routingKey := "test.key"
	messenger := NewRabbitMessenger(conn, exchangeName)

	message := Message{Env: map[string]string{
		"MAIL_TO":       "daniel.rees@autodata.net",
		"RELEASE_LEVEL": "20",
	}}

	err = messenger.Send(message, routingKey, randomString(32))

	if err != nil {
		log.Fatal("Failed to send a message to the queue", err)
	}

	http.HandleFunc("/message", func(writer http.ResponseWriter, request *http.Request) {
		var msg Message
		err := json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			http.Error(writer,err.Error(),400)
			return
		}
		fmt.Fprintf(writer,"Received message: %v",msg)
	})

	http.ListenAndServe(":8002",nil)
}