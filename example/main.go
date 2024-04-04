package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	prettylogger "github.com/rdbell/echo-pretty-logger"
)

const (
	Kilobyte = 1024
	Megabyte = 1024 * Kilobyte
)

func main() {
	// Set up a new Echo server
	e := echo.New()

	// Hello World!
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Redirect example
	e.GET("/redirect", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/")
	})

	// Unauthorized example
	e.GET("/unauthorized", func(c echo.Context) error {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	})

	// POST example
	e.POST("/post", func(c echo.Context) error {
		// Return a 1MB response
		bb := make([]byte, Megabyte)

		return c.Blob(http.StatusOK, "application/octet-stream", bb)
	})

	// Start server
	go func() {
		err := e.Start(":8080")
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for server to start
	for {
		_, err := http.Get("http://localhost:8080/")
		if err == nil {
			break
		}
	}

	// Conditional Logger Middleware
	usePrettyLogger := false
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Condition to decide which logger to use
			if usePrettyLogger {
				return prettylogger.Logger(next)(c)
			}

			return middleware.Logger()(next)(c)
		}
	})

	// Default logger
	log.Println("\n\n\nBefore:")
	testRequests()

	// PettyLogger
	log.Println("\n\n\nAfter:")
	usePrettyLogger = true
	testRequests()
}

func makeRequest(client *http.Client, method, url string) error {
	var request *http.Request
	var err error

	// Handle different request methods
	switch method {
	case http.MethodPost:
		// Generate a byte slice to use as the body of the POST request
		data := make([]byte, Megabyte)
		if _, err := rand.Read(data); err != nil {
			return fmt.Errorf("failed to generate data for POST: %w", err)
		}

		// Use the byte slice as the body of the POST request
		dataReader := bytes.NewReader(data)
		request, err = http.NewRequest(http.MethodPost, url, dataReader)
		if err != nil {
			return fmt.Errorf("failed to create POST request: %w", err)
		}
	case http.MethodGet, http.MethodConnect:
		request, err = http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create %s request: %w", method, err)
		}
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send %s request: %w", method, err)
	}
	defer response.Body.Close()

	// Read and discard the response body
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		return fmt.Errorf("failed to read the response body: %w", err)
	}

	return nil
}

func testRequests() {
	// Setup client
	client := &http.Client{}

	// List of URLs to test along with their request method
	requests := map[string]string{
		"http://localhost:8080/":             http.MethodGet,
		"http://localhost:8080/redirect":     http.MethodGet,
		"http://localhost:8080/unauthorized": http.MethodGet,
		"http://localhost:8080/not_found":    http.MethodConnect,
		"http://localhost:8080/post":         http.MethodPost,
	}

	// Iterate over the requests and make them
	for url, method := range requests {
		err := makeRequest(client, method, url)
		if err != nil {
			log.Fatalf("Failed to make request to %s: %v", url, err)
		}
	}
}
