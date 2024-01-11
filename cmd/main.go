package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"gopkg.in/redis.v5"
)

var (
	kafkaProducer sarama.SyncProducer
	redisClient   *redis.Client
	mu            sync.Mutex
)

func main() {
	// Setup Kafka producer
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	kafkaConfig.Producer.Return.Successes = true                  // Enable success notifications

	var err error
	kafkaProducer, err = sarama.NewSyncProducer([]string{"localhost:29092"}, kafkaConfig)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}

	// Setup Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// Setup HTTP server
	router := gin.Default()

	router.POST("/pushData", pushDataHandler)
	router.GET("/getData", getDataHandler)

	err = router.Run(":8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func pushDataHandler(c *gin.Context) {
	var requestData map[string]interface{}

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	log.Printf("Invalid Kafka message: %s", requestData)
	// Goroutine-1: Route to Kafka
	go routeToKafka(requestData)
	c.JSON(http.StatusOK, gin.H{"message": "Data received and pushed to Kafka"})
}

func getDataHandler(c *gin.Context) {
	// Goroutine-2: Route from Kafka to Redis
	go routeFromKafkaToRedis()

	// Fetch data from Redis and Kafka, and deliver the response
	data, err := fetchDataFromRedis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data from Redis"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func routeToKafka(data map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	// Produce the data to Kafka
	message := &sarama.ProducerMessage{
		Topic: "messagetopic",
		Value: sarama.StringEncoder(fmt.Sprintf("%v", data)),
	}
	log.Printf("Invalid Kafka message: %s", message.Value)
	_, _, err := kafkaProducer.SendMessage(message)
	if err != nil {
		log.Printf("Error producing to Kafka: %v", err)
	}
}

func routeFromKafkaToRedis() {
	// Consume messages from Kafka
	consumer, err := sarama.NewConsumer([]string{"localhost:29092"}, nil)
	if err != nil {
		log.Printf("Error creating Kafka consumer: %v", err)
		return
	}
	//defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("messagetopic", 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Error creating partition consumer: %v", err)
		return
	}
	//defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			// Goroutine-2: Process Kafka message and route to Redis
			go processKafkaMessage(msg)
		}
	}
}

func processKafkaMessage(msg *sarama.ConsumerMessage) {
	var data map[string]interface{}

	// Assuming the message is a map
	err := json.Unmarshal(msg.Value, &data)
	if err != nil {
		log.Printf("Error unmarshalling Kafka message: %v", err)
		log.Printf("Invalid Kafka message: %s", string(msg.Value))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Store data in Redis
	err = storeDataInRedis(data)
	if err != nil {
		log.Printf("Error storing data in Redis: %v", err)
	}
}

func storeDataInRedis(data map[string]interface{}) error {
	// Store data in Redis
	// Assuming data has "uid" and "message" fields
	uid := fmt.Sprintf("%v", data["uid"])
	message := fmt.Sprintf("%v", data["message"])

	err := redisClient.HSet("message:", uid, message).Err()
	return err
}

func fetchDataFromRedis() ([]map[string]string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Fetch data from Redis
	redisData, err := redisClient.HGetAll("message:").Result()
	if err != nil {
		return nil, err
	}

	// Convert Redis data to the desired format
	var result []map[string]string
	for key, value := range redisData {
		result = append(result, map[string]string{"uid": key, "message": value})
	}

	return result, nil
}