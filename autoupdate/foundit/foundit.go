package foundit

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const attachResumeURL = "https://www.foundit.in/seeker-profile/api/attachResume?saveToDB=false&isExperienced=true"

// Config holds FoundIt-specific configuration
type Config struct {
	CookieFile string
}

// Run executes the full FoundIt resume update workflow
func Run(filePath string, cfg Config) error {
	if cfg.CookieFile == "" {
		return fmt.Errorf("cookie file is required for FoundIt")
	}

	cookies, err := readCookieHeader(cfg.CookieFile)
	if err != nil {
		return fmt.Errorf("failed to read cookie file: %w", err)
	}

	if !isAuthenticated(cookies) {
		return fmt.Errorf("missing required authentication cookies (MSSOAT, MSAL)")
	}

	resp, body, err := uploadResume(filePath, cookies)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("foundit: upload status: %s\n", resp.Status)
	fmt.Printf("foundit: response: %s\n", string(body))

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
	// FoundIt uses MSSOAT and MSAL cookies for authentication
	if _, ok := cookies["MSSOAT"]; !ok {
		return false
	}
	if _, ok := cookies["MSAL"]; !ok {
		return false
	}
	return true
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

func buildCookieHeader(cookies map[string]string) string {
	keys := make([]string, 0, len(cookies))
	for name := range cookies {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", name, cookies[name]))
	}
	return strings.Join(parts, "; ")
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

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, attachResumeURL, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Referer", "https://www.foundit.in/seeker/profile")
	req.Header.Set("Sec-CH-UA", `"Chromium";v="148", "Google Chrome";v="148", "Not/A)Brand";v="99"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
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
