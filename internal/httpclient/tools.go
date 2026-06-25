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

// httpClient dùng chung với timeout 30 giây - thực hành tốt trong production.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

const maxResponseSize = 1 * 1024 * 1024 // 1MB

// Get thực hiện HTTP GET request.
// Ứng dụng thực tế: Gọi REST API nội bộ, lấy dữ liệu từ webhook, health check.
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
		return "", fmt.Errorf("không thể đọc response body: %w", err)
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

// PostJSON thực hiện HTTP POST với body JSON.
// Ứng dụng thực tế: Gửi dữ liệu đến API, trigger webhook, tạo resource qua REST.
func PostJSON(rawURL, jsonBody string) (string, error) {
	if err := validateURL(rawURL); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBufferString(jsonBody))
	if err != nil {
		return "", fmt.Errorf("không thể tạo request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request thất bại: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", fmt.Errorf("không thể đọc response body: %w", err)
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

// validateURL kiểm tra URL hợp lệ và chỉ cho phép http/https (chống SSRF).
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL không hợp lệ: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("chỉ hỗ trợ giao thức http và https, nhận được: '%s'", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL thiếu host")
	}
	return nil
}
