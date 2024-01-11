// main.go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
    kafkaProducer sarama.AsyncProducer
    kafkaConsumer sarama.Consumer
    redisClient   *redis.Client
    wg            sync.WaitGroup
)

func main() {
    initializeKafkaProducer()
    initializeKafkaConsumer()
    initializeRedisClient()

    router := mux.NewRouter()

    router.HandleFunc("/pushdata", pushDataHandler).Methods("POST")
    router.HandleFunc("/getdata", getDataHandler).Methods("GET")

    port := 8080
    fmt.Printf("Server is running on port %d...\n", port)
    http.ListenAndServe(fmt.Sprintf(":%d", port), router)

    // Close resources when the server shuts down
    defer func() {
        kafkaProducer.Close()
        kafkaConsumer.Close()
        redisClient.Close()
    }()
}

func generateUID() string {
    // Implement a function to generate a UID (e.g., using a UUID library)
    // Example: Using the `github.com/google/uuid` package
    uid, err := uuid.NewRandom()
    if err != nil {
        panic(err)
    }
    return uid.String()
}
func initializeKafkaProducer() {
    // Initialize Kafka producer
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    config.Producer.Return.Errors = true

    p, err := sarama.NewAsyncProducer([]string{"localhost:29092"}, config)
    if err != nil {
        panic(err)
    }
    kafkaProducer = p

    // Start a goroutine to handle successful and failed deliveries
    wg.Add(1)
    go handleKafkaProducerEvents()
}

func initializeKafkaConsumer() {
    // Initialize Kafka consumer
    config := sarama.NewConfig()
    config.Consumer.Return.Errors = true

    c, err := sarama.NewConsumer([]string{"localhost:29092"}, config)
    if err != nil {
        panic(err)
    }
    kafkaConsumer = c

    // Start a goroutine to consume messages from Kafka
    wg.Add(1)
    go consumeFromKafka()
}

func initializeRedisClient() {
    // Initialize Redis client
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        // Add other Redis configuration options if needed
    })
}

func pushDataHandler(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }
	uid := generateUID()
    // Produce message to Kafka topic in a separate goroutine
    wg.Add(1)
    go produceToKafka(uid, body)

    fmt.Fprint(w, "Data received successfully")
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {
    // Retrieve data from Redis and send it to Kafka in a separate goroutine
    wg.Add(1)
    go retrieveDataAndSendToKafka(w)
}

func retrieveDataAndSendToKafka(w http.ResponseWriter) {
	defer wg.Done()

    // Retrieve all data from Redis
    ctx := context.Background()
    keys, err := redisClient.Keys(ctx, "*").Result()
    if err != nil {
        http.Error(w, "Error retrieving data from Redis", http.StatusInternalServerError)
        return
    }

    // Send each data to Kafka
    for _, key := range keys {
        data, err := redisClient.Get(ctx, key).Result()
        if err != nil {
            fmt.Printf("Error retrieving data with key %s from Redis: %v\n", key, err)
            continue
        }

        // Send the retrieved data to Kafka
        wg.Add(1)
        go produceToKafka(key, []byte(data))
    }

    // Deliver the response
    fmt.Fprint(w, "All data retrieved and sent to Kafka successfully")
}

func produceToKafka(uid string, data []byte) {
    defer wg.Done()

    // Produce message to Kafka topic
    message := &sarama.ProducerMessage{
        Topic: "testTopic",
		Key: sarama.StringEncoder(uid),
        Value: sarama.StringEncoder(data),
    }

    // Asynchronously send the message to Kafka
    kafkaProducer.Input() <- message
}

func handleKafkaProducerEvents() {
    defer wg.Done()

    for {
        select {
        case success := <-kafkaProducer.Successes():
            fmt.Printf("Produced message to Kafka: %v\n", success)
        case err := <-kafkaProducer.Errors():
            fmt.Printf("Error producing message to Kafka: %v\n", err.Err)
        }
    }
}

func consumeFromKafka() {
    defer wg.Done()

    partitionConsumer, err := kafkaConsumer.ConsumePartition("testTopic", 0, sarama.OffsetOldest)
    if err != nil {
        panic(err)
    }

    defer func() {
        if err := partitionConsumer.Close(); err != nil {
            panic(err)
        }
    }()

    for {
        // Consume messages from Kafka topic
        msg := <-partitionConsumer.Messages()

        // Process the received data
        fmt.Printf("Received message from Kafka: %s\n", string(msg.Value))

        // Route the data to Redis in a separate goroutine
        wg.Add(1)
        go routeToRedis(string(msg.Key),msg.Value)
    }
}

func routeToRedis(uid string, data []byte) {
    defer wg.Done()

    // Route data to Redis
    ctx := context.Background()
    err := redisClient.Set(ctx, uid, string(data), 0).Err()
    if err != nil {
        fmt.Printf("Error routing data to Redis: %v\n", err)
    }
}

func retrieveAllDataAndSendToKafka(w http.ResponseWriter) {
    defer wg.Done()

    // Retrieve all data from Redis
    ctx := context.Background()
    keys, err := redisClient.Keys(ctx, "*").Result()
    if err != nil {
        http.Error(w, "Error retrieving data from Redis", http.StatusInternalServerError)
        return
    }

    // Send each data to Kafka
    for _, key := range keys {
        data, err := redisClient.Get(ctx, key).Result()
        if err != nil {
            fmt.Printf("Error retrieving data with key %s from Redis: %v\n", key, err)
            continue
        }

        // Send the retrieved data to Kafka
        wg.Add(1)
        go produceToKafka(key, []byte(data))
    }

    // Deliver the response
    fmt.Fprint(w, "All data retrieved and sent to Kafka successfully")
}