package scale

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var workerCount = 5

type Job struct {
	ID int
}

func Main() {
	//Creating a buffer channel to control the number of concurrent workers
	jobQueue := make(chan Job, 10)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobQueue, &wg)
	}

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		job := Job{ID: time.Now().Nanosecond()}
		go func ()  {
			jobQueue <- job
		}()
		fmt.Fprint(w, "Jo queued\n")
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("Server error: ", err)
		}
	}()

	fmt.Println("Server is running on http://localhost:8080")

	// Simulate client sending a lot of requests
	for i := 0; i< 10; i++ {
		go sendRequest()
	}

	wg.Wait()

	close(jobQueue)
}

func worker(id int, jobQueue <- chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobQueue {
		// Simulate a compute-heavy operation by sleeping for a duration
		time.Sleep(2 * time.Second)

		// Log the completion of the job
		log.Printf("Worker %d completed job %d\n", id, job.ID)
	}
}

func sendRequest() {
	// Simulate sending a client request to the server
	resp, err := http.Get("http://localhost:8080/process")
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()
	log.Println("Response:", resp.Status)
}