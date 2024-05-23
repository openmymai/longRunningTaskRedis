package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/openmymai/longRunningTaskRedis/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type JobServer struct {
	Queue       amqp.Queue
	Channel     *amqp.Channel
	Conn        *amqp.Connection
	redisClient *redis.Client
}

func (s *JobServer) publish(jsonBody []byte) error {
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBody,
	}
	err := s.Channel.Publish(
		"",        // exchange
		queueName, // routing key(Queue)
		false,     // mandatory
		false,     // immediate
		message,
	)
	handleError(err, "Error while generating JobID")
	return err
}

func (s *JobServer) asyncDBHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()
	queryParams := r.URL.Query()
	// Ex: client_time: 1569174071
	unixTime, err := strconv.ParseInt(queryParams.Get("client_time"), 10, 64)
	clientTime := time.Unix(unixTime, 0)
	handleError(err, "Error while converting client time")

	jsonBody, err := json.Marshal(models.Job{
		ID:        jobID,
		Type:      "A",
		ExtraData: models.Log{ClientTime: clientTime},
	})
	handleError(err, "JSON body creation failed")

	if s.publish(jsonBody) == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBody)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *JobServer) asyncCallbackHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()

	jsonBody, err := json.Marshal(models.Job{
		ID:        jobID,
		Type:      "B",
		ExtraData: models.CallBack{CallBackURL: "example.org"},
	})
	handleError(err, "JSON body creation failed")

	if s.publish(jsonBody) == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBody)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *JobServer) asyncMailHandler(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.NewRandom()

	jsonBody, err := json.Marshal(models.Job{
		ID:        jobID,
		Type:      "C",
		ExtraData: models.Mail{EmailAddress: "packt@example.org"},
	})
	handleError(err, "JSON body creation failed")

	if s.publish(jsonBody) == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBody)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getServer(name string) JobServer {
	/*
		Creates a server object and initiates
		the Channel and Queue details to publish messages
	*/
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	handleError(err, "Dialing failed to RabbitMQ broker")
	channel, err := conn.Channel()
	handleError(err, "Fetching channel failed")
	jobQueue, err := channel.QueueDeclare(
		name,  // Name of the queue
		false, // Message is persisted or not
		false, // Delete message when unused
		false, // Exclusive
		false, // No Waiting time
		nil,   // Extra args
	)
	handleError(err, "Job queue creation failed")
	return JobServer{Conn: conn, Channel: channel, Queue: jobQueue}
}

func (s *JobServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	// fetch UUID from query
	uuid := queryParams.Get("uuid")
	w.Header().Set("Content-Type", "application/json")
	jobStatus := s.redisClient.Get(uuid)
	status := map[string]string{"uuid": uuid, "status": jobStatus.Val()}
	response, err := json.Marshal(status)
	handleError(err, "Cannot create response for client")
	w.Write(response)
}
