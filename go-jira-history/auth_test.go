package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/andygrunwald/go-jira"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	url := os.Getenv("JIRA_URL")
	token := os.Getenv("JIRA_PERSONAL_ACCESS_TOKEN")

	// Print values (except token) for debugging
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Token length: %d\n", len(token))

	// Create a custom HTTP client with request logging
	httpClient := &http.Client{
		Transport: &loggingTransport{
			transport: http.DefaultTransport,
		},
	}

	// Create JIRA client with token auth
	tp := jira.PATAuthTransport{
		Token: token,
		Transport: &customTransport{
			base: httpClient.Transport,
			headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
		},
	}

	client, err := jira.NewClient(tp.Client(), url)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// Try to get current user
	user, resp, err := client.User.GetSelf()
	if err != nil {
		if resp != nil {
			log.Printf("Response Status: %s", resp.Response.Status)
			log.Printf("Response Status Code: %d", resp.Response.StatusCode)
			
			// Log response headers for debugging
			log.Printf("Response Headers:")
			for k, v := range resp.Response.Header {
				log.Printf("%s: %v", k, v)
			}
		}
		log.Fatalf("Error authenticating: %v", err)
	}

	fmt.Printf("Successfully authenticated as: %s\n", user.DisplayName)
}

// loggingTransport is a custom transport that logs requests and responses
type loggingTransport struct {
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	reqBody, _ := httputil.DumpRequestOut(req, true)
	log.Printf("Request: %s", string(reqBody))

	// Perform the request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log the response
	respBody, _ := httputil.DumpResponse(resp, true)
	log.Printf("Response: %s", string(respBody))

	return resp, nil
}

// customTransport adds custom headers to requests
type customTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
} 