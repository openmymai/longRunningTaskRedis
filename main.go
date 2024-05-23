package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName string = "jobQueue"
const hostString string = "127.0.0.1:8000"

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	jobServer := getServer(queueName)

	jobServer.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	go func(conn *amqp.Connection) {
		workerProcess := Workers{
			conn: jobServer.Conn,
		}
		workerProcess.run()
	}(jobServer.Conn)

	router := mux.NewRouter()

	router.HandleFunc("/job/database", jobServer.asyncDBHandler)
	router.HandleFunc("/job/mail", jobServer.asyncMailHandler)
	router.HandleFunc("/job/callback", jobServer.asyncCallbackHandler)
	router.HandleFunc("/job/status", jobServer.statusHandler)

	httpServer := &http.Server{
		Handler:      router,
		Addr:         hostString,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(httpServer.ListenAndServe())

	defer jobServer.Channel.Close()
	defer jobServer.Conn.Close()
}
