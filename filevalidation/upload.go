package filevalidation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	uploadURL         = "https://filevalidation.naukri.com/file"
	loginStatusURL    = "https://www.naukri.com/central-login-services/v0/credentials/login-status"
	submitURLTemplate = "https://www.naukri.com/cloudgateway-mynaukri/resman-aggregator-services/v0/users/self/profiles/%s/advResume"
)

// UploadFile uploads a local PDF file to the Naukri file validation endpoint.
// formKey and fileKey are passed as form fields to match the curl payload.
func UploadFile(filePath, formKey, fileKey string) (*http.Response, []byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.SetBoundary("----WebKitFormBoundary49FDqx5wy6ri9aYU"); err != nil {
		return nil, nil, fmt.Errorf("set boundary: %w", err)
	}

	if err := writer.WriteField("formKey", formKey); err != nil {
		return nil, nil, fmt.Errorf("add formKey: %w", err)
	}

	fileHeader := textproto.MIMEHeader{}
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filepath.Base(filePath)))
	fileHeader.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(fileHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("create file part: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, nil, fmt.Errorf("copy file contents: %w", err)
	}

	if err := writer.WriteField("fileName", filepath.Base(filePath)); err != nil {
		return nil, nil, fmt.Errorf("add fileName: %w", err)
	}
	if err := writer.WriteField("uploadCallback", "true"); err != nil {
		return nil, nil, fmt.Errorf("add uploadCallback: %w", err)
	}
	if err := writer.WriteField("fileKey", fileKey); err != nil {
		return nil, nil, fmt.Errorf("add fileKey: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-IN,en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Appid", "105")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Origin", "https://www.naukri.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://www.naukri.com/")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", "\"macOS\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Systemid", "fileupload")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("perform request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return resp, nil, fmt.Errorf("read response: %w", err)
	}

	return resp, respBody, nil
}

func ReadCookieHeader(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read cookie file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func SaveCookieHeader(filePath, cookieHeader string) error {
	return os.WriteFile(filePath, []byte(cookieHeader), 0644)
}

func IsTokenValid(cookieHeader string) (token string, valid bool) {
	cookies := parseCookieHeader(cookieHeader)
	token, exists := cookies["nauk_at"]
	if !exists || token == "" {
		return "", false
	}
	return token, true
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

func mergeCookieHeader(original string, setCookies []*http.Cookie) string {
	cookieMap := parseCookieHeader(original)
	for _, c := range setCookies {
		if c.Name == "" {
			continue
		}
		cookieMap[c.Name] = c.Value
	}
	return buildCookieHeader(cookieMap)
}

func parseSetCookieHeader(header string) *http.Cookie {
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return nil
	}
	pair := strings.TrimSpace(parts[0])
	kv := strings.SplitN(pair, "=", 2)
	if len(kv) != 2 {
		return nil
	}
	return &http.Cookie{Name: kv[0], Value: kv[1]}
}

func parseSetCookies(headers []string) []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(headers))
	for _, header := range headers {
		if c := parseSetCookieHeader(header); c != nil {
			cookies = append(cookies, c)
		}
	}
	return cookies
}

// LoginStatus calls the Naukri login-status endpoint using cookies from a file,
// then returns the bearer access token from the returned nauk_at cookie.
func LoginStatus(cookieHeader string) (bearerToken, mergedCookieHeader string, err error) {
	req, err := http.NewRequest(http.MethodGet, loginStatusURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create login-status request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-IN,en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Appid", "105")
	req.Header.Set("Content-Type", "application/json")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
	req.Header.Set("Origin", "https://www.naukri.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://www.naukri.com/mnjuser/profile")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", "\"macOS\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Systemid", "jobseeker")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("perform login-status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("login-status returned %s: %s", resp.Status, string(body))
	}

	setCookies := parseSetCookies(resp.Header["Set-Cookie"])
	mergedCookieHeader = mergeCookieHeader(cookieHeader, setCookies)

	for _, c := range setCookies {
		if c.Name == "nauk_at" {
			bearerToken = c.Value
			break
		}
	}

	if bearerToken == "" {
		originalCookies := parseCookieHeader(cookieHeader)
		bearerToken = originalCookies["nauk_at"]
	}

	if bearerToken == "" {
		return "", mergedCookieHeader, fmt.Errorf("nauk_at token not found in login-status response")
	}

	return bearerToken, mergedCookieHeader, nil
}

type submitRequest struct {
	TextCV struct {
		FormKey       string      `json:"formKey"`
		FileKey       string      `json:"fileKey"`
		TextCvContent interface{} `json:"textCvContent"`
	} `json:"textCV"`
}

// SubmitResume sends the uploaded resume metadata to Naukri's advResume endpoint.
func SubmitResume(profileID, formKey, fileKey, bearerToken, cookieHeader string) (*http.Response, []byte, error) {
	url := fmt.Sprintf(submitURLTemplate, profileID)
	reqBody := submitRequest{}
	reqBody.TextCV.FormKey = formKey
	reqBody.TextCV.FileKey = fileKey
	reqBody.TextCV.TextCvContent = nil

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("create submit request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-IN,en-GB;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Appid", "105")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	req.Header.Set("Content-Type", "application/json")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
	req.Header.Set("Origin", "https://www.naukri.com")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://www.naukri.com/mnjuser/profile?id=&altresid")
	req.Header.Set("Sec-CH-UA", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", "\"macOS\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Systemid", "105")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	req.Header.Set("X-HTTP-Method-Override", "PUT")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("perform submit request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return resp, nil, fmt.Errorf("read submit response: %w", err)
	}

	return resp, respBody, nil
}
