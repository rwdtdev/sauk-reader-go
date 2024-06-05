package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Global variable to manage file lock
var fileMutex = &sync.Mutex{}

// Initializes the server by reading environment variables, setting up the TCP
// listener, and starting the retry mechanism for failed requests.
func main() {
	// Retrieve environment variables
	listenPort := os.Getenv("LISTEN_PORT")
	endpointURL := os.Getenv("ENDPOINT_URL")
	retryFile := os.Getenv("RETRY_FILE")

	// Validate required environment variables
	if listenPort == "" || endpointURL == "" {
		log.Fatal("LISTEN_PORT and ENDPOINT_URL environment variables must be set")
	}
	if retryFile == "" {
		log.Fatal("RETRY_FILE environment variable must be set")
	}

	// Listen on the specified TCP port
	listener, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", listenPort, err)
	}
	defer listener.Close()

	log.Printf("Server is listening on TCP port %s", listenPort)
	log.Printf("Forwarding data to %s", endpointURL)

	// Start the retry mechanism in a separate goroutine
	go retryFailedRequests(endpointURL, retryFile)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each connection in a separate goroutine
		go handleConnection(conn, endpointURL, retryFile)
	}
}

// handleConnection processes the incoming connection and forwards data to the endpoint
func handleConnection(conn net.Conn, endpointURL, retryFile string) {
	defer conn.Close()

	// Read the incoming data
	buffer, err := io.ReadAll(conn)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	log.Printf("Received data: %s", buffer)

	// Split the buffer into lines
	lines := bytes.Split(buffer, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Unmarshal the JSON data into a map for each line
		var dataMap map[string]interface{}
		err = json.Unmarshal(line, &dataMap)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			continue
		}

		// Add the current timestamp
		dataMap["Timestamp"] = time.Now().Format(time.RFC3339)

		// Marshal the modified data back to JSON
		modifiedData, err := json.Marshal(dataMap)
		if err != nil {
			log.Printf("Error marshalling modified JSON: %v", err)
			continue
		}

		// Forward the modified data as a POST request to the endpoint URL
		resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(modifiedData))
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("Error forwarding request, saving data for retry: %v", err)
			saveDataForRetry(modifiedData, retryFile)
			continue
		}

		defer resp.Body.Close()
	}
}

// saveDataForRetry saves the data to a file for future retries in case of failure
func saveDataForRetry(data []byte, retryFile string) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(retryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		log.Printf("Error writing to file: %v", err)
		return
	}

	// Write a newline to separate data entries
	_, err = file.WriteString("\n")
	if err != nil {
		log.Printf("Error writing newline to file: %v", err)
	}
}

// retryFailedRequests periodically retries sending failed requests saved in the retry file
func retryFailedRequests(endpointURL, retryFile string) {
	for {
		fileMutex.Lock()

		// Read the file
		file, err := os.OpenFile(retryFile, os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			fileMutex.Unlock()
			continue
		}

		scanner := bufio.NewScanner(file)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		// Clear the file content after reading
		err = os.Truncate(retryFile, 0)
		fileMutex.Unlock()

		if err != nil {
			log.Printf("Error truncating file: %v", err)
			continue
		}

		// Try to resend each line
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Resend the data
			resp, err := http.Post(endpointURL, "application/json", bytes.NewBufferString(line))
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("Error re-forwarding request, saving data for retry: %v", err)
				saveDataForRetry([]byte(line), retryFile)
			}
		}

		// Wait for some time before retrying
		time.Sleep(1 * time.Minute)
	}
}
