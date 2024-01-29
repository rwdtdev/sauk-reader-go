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

func main() {
	listenPort := os.Getenv("LISTEN_PORT")
	endpointURL := os.Getenv("ENDPOINT_URL")
	retryFile := os.Getenv("RETRY_FILE")

	if listenPort == "" || endpointURL == "" {
		log.Fatal("LISTEN_PORT and ENDPOINT_URL environment variables must be set")
	}
	if retryFile == "" {
		log.Fatal("RETRY_FILE environment variable must be set")
	}

	// Listen on TCP port
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
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn, endpointURL, retryFile)
	}
}

func handleConnection(conn net.Conn, endpointURL, retryFile string) {
	defer conn.Close()

	// Read the incoming data
	buffer, err := io.ReadAll(conn)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	log.Printf("Received data: %s", buffer)

	// Unmarshal the JSON data into a map
	var dataMap map[string]interface{}
	err = json.Unmarshal(buffer, &dataMap)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return
	}

	// Add the current timestamp
	dataMap["Timestamp"] = time.Now().Format(time.RFC3339)

	// Marshal the modified data back to JSON
	modifiedData, err := json.Marshal(dataMap)
	if err != nil {
		log.Printf("Error marshalling modified JSON: %v", err)
		return
	}

	// Forward the modified data as a POST request to the endpoint URL
	// Attempt to send the data
	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(modifiedData))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Error forwarding request, saving data for retry: %v", err)
		saveDataForRetry(modifiedData, retryFile)
		return
	}

	defer resp.Body.Close()
}

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
