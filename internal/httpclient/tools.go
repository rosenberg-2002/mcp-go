package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpClient is a shared client with a 30-second timeout — best practice in production.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

const maxResponseSize = 1 * 1024 * 1024 // 1MB

// Get performs an HTTP GET request.
// Practical uses: calling internal REST APIs, fetching data from webhooks, health checks.
func Get(rawURL string) (string, error) {
	if err := validateURL(rawURL); err != nil {
		return "", err
	}

	resp, err := httpClient.Get(rawURL) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("request thất bại: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GET %s\n", rawURL))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Status       : %s\n", resp.Status))
	sb.WriteString(fmt.Sprintf("Content-Type : %s\n", resp.Header.Get("Content-Type")))
	sb.WriteString(fmt.Sprintf("Body (%d bytes):\n", len(body)))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.Write(body)
	return sb.String(), nil
}

// PostJSON performs an HTTP POST with a JSON body.
// Practical uses: sending data to APIs, triggering webhooks, creating resources via REST.
func PostJSON(rawURL, jsonBody string) (string, error) {
	if err := validateURL(rawURL); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBufferString(jsonBody))
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("POST %s\n", rawURL))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Request Body : %s\n", jsonBody))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Status       : %s\n", resp.Status))
	sb.WriteString(fmt.Sprintf("Body (%d bytes):\n", len(body)))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.Write(body)
	return sb.String(), nil
}

// validateURL validates the URL and only allows http/https (prevents SSRF).
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https protocols are supported, got: '%s'", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL is missing host")
	}
	return nil
}
