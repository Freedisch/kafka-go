# Webserver with Messaging System Integration

## Problem Statement

The task is to build a web server in Go to handle incoming data and route it through messaging systems. The web server should have two endpoints: one to receive data via POST requests and another to retrieve data via GET requests.

### Specifications:

1. **POST Endpoint:**
   - Receives incoming data through a POST request.
   - Utilizes Goroutine-1 to route the data to a Kafka topic.
   - Another Goroutine-2 listens to this Kafka topic and routes the data to a Redis database.

2. **GET Endpoint:**
   - Retrieves data from Redis through a GET request.
   - Uses Goroutine-3 to send the data to Kafka and deliver the response.

## Design of Request/Response

- **POST Request Flow:**
  1. Client sends a POST request to the designated endpoint.
  2. Web server handles the request, spawns Goroutine-1 to send data to Kafka topic.
  3. Goroutine-2 picks up the data from the Kafka topic and routes it to Redis.

- **GET Request Flow:**
  1. Client sends a GET request to the designated endpoint.
  2. Web server retrieves data from Redis using Goroutine-3.
  3. The retrieved data is sent to Kafka for further processing, and the response is delivered to the client.

## Implementation Details

- Use a local Docker-based Kafka system for testing and development.
- Utilize the Sarama library, a reliable Kafka client for Go.
- Ensure that transactions are non-blocking to prevent issues with Golang concurrency.

## Setting Up the Development Environment

1. Install Docker: [Docker Installation Guide](https://docs.docker.com/get-docker/)
2. Set up a local Kafka system using Docker.
   ```bash
   docker run -d --name kafka -p 9092:9092 -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT -e KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT wurstmeister/kafka
   ```
3. Use the Sarama library for Go. Install it using:
   ```bash
   go get -u github.com/Shopify/sarama
   ```

## Notes and Considerations

- Sarama is recommended for Kafka integration due to its reliability and widespread use in the Go community.
- Transactions in Golang should be non-blocking to maintain the efficiency of the web server.

---

# Simple Scalable Webserver

## Problem Statement

Build a simple web server in Go that scales as it receives more requests. The server should spawn more Goroutines as the same endpoint handles numerous requests.

### Specifications:

- **Endpoint:**
  - The application should have only one endpoint.
  - Introduce a sleep operation to simulate a compute-heavy operation.
  - The server should scale by spawning additional Goroutines as requests increase within the sleep interval.

## Solution Approach

- Study and implement a mechanism to dynamically scale Goroutines based on incoming requests.
- Introduce sleep operations to simulate compute-heavy tasks and test the scalability of the server.
- Design the system to handle an increasing number of requests efficiently.

## Implementation Details

- Implement a single endpoint for the web server.
- Utilize Goroutines to handle incoming requests concurrently.
- Introduce sleep operations within the Goroutines to simulate compute-heavy tasks.
- Monitor and scale Goroutines dynamically based on incoming request patterns.

## Notes and Considerations

- The scalability of the server should be tested under various load conditions.
- Fine-tune the sleep interval to ensure realistic simulation of compute-heavy tasks.

---
