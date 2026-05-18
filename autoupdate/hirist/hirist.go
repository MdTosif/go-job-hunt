package hirist

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const uploadURL = "https://gladiator.hirist.tech/media/files?pref=kp_br_prm&refPool=%7B%22pref%22:%22kp_br_prm%22%7D"

// Config holds Hirist-specific configuration.
type Config struct {
	CookieFile string
}

// Run executes the Hirist resume upload workflow.
func Run(filePath string, cfg Config) error {
	if cfg.CookieFile == "" {
		return fmt.Errorf("cookie file is required for Hirist")
	}

	cookieHeader, err := readCookieHeader(cfg.CookieFile)
	if err != nil {
		return fmt.Errorf("failed to read cookie file: %w", err)
	}

	if !isAuthenticated(cookieHeader) {
		return fmt.Errorf("missing required authentication cookies (HIRIST_CK1 or hirist_seeker_enc)")
	}

	resp, body, err := uploadResume(filePath, cookieHeader)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("hirist: upload status: %s\n", resp.Status)
	fmt.Printf("hirist: response: %s\n", string(body))

	return nil
}

func readCookieHeader(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read cookie file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func isAuthenticated(cookieHeader string) bool {
	cookies := parseCookieHeader(cookieHeader)

	if cookies["HIRIST_CK1"] != "" {
		return true
	}
	if cookies["hirist_seeker_enc"] != "" {
		return true
	}

	return false
}

func parseCookieHeader(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		cookies[kv[0]] = kv[1]
	}

	return cookies
}

func uploadResume(filePath, cookieHeader string) (*http.Response, []byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, nil, fmt.Errorf("copy file contents: %w", err)
	}

	if err := writer.WriteField("mediaType", "TextResume"); err != nil {
		return nil, nil, fmt.Errorf("add mediaType: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-IN,en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Origin", "https://www.hirist.tech")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://www.hirist.tech/")
	req.Header.Set("Sec-CH-UA", `"Chromium";v="148", "Google Chrome";v="148", "Not/A)Brand";v="99"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", cookieHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("perform request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("read response: %w", err)
	}

	return resp, respBody, nil
}
