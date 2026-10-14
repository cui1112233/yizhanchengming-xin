package videomanagementconnector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TargetHTTPService is the Go-owned transport for the video-management site.
// It deliberately contains no browser or Node dependency. A CDP implementation
// can be added behind the same Service interface for sites that require a UI.
type TargetHTTPService struct {
	BaseURL string
	Client  *http.Client
}

func (s *TargetHTTPService) Login(ctx context.Context, input LoginInput) (Session, error) {
	base, err := s.base()
	if err != nil {
		return Session{}, err
	}
	client := s.client()
	landing, err := s.do(ctx, client, http.MethodGet, base+"/tttadmin/login.php", "", nil)
	if err != nil {
		return Session{}, fmt.Errorf("target site login page: %w", err)
	}
	initial := mergeCookies(landing.Header.Values("Set-Cookie"))
	landing.Body.Close()
	payload, _ := json.Marshal(input)
	response, err := s.do(ctx, client, http.MethodPost, base+"/tttadmin/api/login.php", initial, bytes.NewReader(payload))
	if err != nil {
		return Session{}, fmt.Errorf("target site login: %w", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Session{}, fmt.Errorf("target site login returned HTTP %d", response.StatusCode)
	}
	var result map[string]any
	if json.Unmarshal(body, &result) != nil || result["success"] != true {
		return Session{}, fmt.Errorf("target site login rejected")
	}
	cookie := mergeCookies(append([]string{initial}, response.Header.Values("Set-Cookie")...))
	if cookie == "" {
		return Session{}, fmt.Errorf("target site did not return a session")
	}
	return Session{Cookie: cookie, TargetUsername: strings.TrimSpace(input.Username)}, nil
}

func (s *TargetHTTPService) RefreshCatalogs(ctx context.Context, session Session) (Catalogs, error) {
	base, err := s.base()
	if err != nil {
		return Catalogs{}, err
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/organization.php", session.Cookie, nil)
	if err != nil {
		return Catalogs{}, fmt.Errorf("target site catalogs: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Catalogs{}, fmt.Errorf("target site catalogs returned HTTP %d", response.StatusCode)
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	return Catalogs{Raw: value}, nil
}

func (s *TargetHTTPService) CheckSession(ctx context.Context, session Session) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/zidingyi.php", session.Cookie, nil)
	if err != nil {
		return fmt.Errorf("target site session: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("target site session expired")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("target site session returned HTTP %d", response.StatusCode)
	}
	return nil
}

func (s *TargetHTTPService) Submit(ctx context.Context, session Session, input SubmissionInput) (Submission, error) {
	base, err := s.base()
	if err != nil {
		return Submission{}, err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range input.FormFields {
		if err := writer.WriteField(key, value); err != nil {
			return Submission{}, err
		}
	}
	if input.BookID != "" {
		_ = writer.WriteField("bookid", input.BookID)
	}
	filename := input.Filename
	if filename == "" {
		filename = input.BookID + ".txt"
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return Submission{}, err
	}
	if _, err := part.Write(input.Content); err != nil {
		return Submission{}, err
	}
	if err := writer.Close(); err != nil {
		return Submission{}, err
	}
	response, err := s.doWithContentType(ctx, s.client(), http.MethodPost, base+"/tttadmin/api/zbooklist_upload.php", session.Cookie, &body, writer.FormDataContentType())
	if err != nil {
		return Submission{}, fmt.Errorf("target site submit: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Submission{}, fmt.Errorf("target site submit returned HTTP %d", response.StatusCode)
	}
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	if result["success"] != true {
		return Submission{}, fmt.Errorf("target site rejected submission")
	}
	id := strings.TrimSpace(fmt.Sprint(result["id"]))
	if id == "<nil>" {
		id = ""
	}
	return Submission{ID: id, Status: "submitted", ReadbackRaw: result}, nil
}

func (s *TargetHTTPService) GetSubmission(ctx context.Context, session Session, id string) (Submission, error) {
	base, err := s.base()
	if err != nil {
		return Submission{}, err
	}
	query := url.Values{"id": []string{strings.TrimSpace(id)}}.Encode()
	response, err := s.do(ctx, s.client(), http.MethodGet, base+"/tttadmin/api/zbooklist.php?"+query, session.Cookie, nil)
	if err != nil {
		return Submission{}, fmt.Errorf("target site readback: %w", err)
	}
	defer response.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Submission{}, fmt.Errorf("target site readback returned HTTP %d", response.StatusCode)
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		value = string(raw)
	}
	return Submission{ID: id, Status: "readback", ReadbackRaw: value}, nil
}

func (s *TargetHTTPService) base() (string, error) {
	raw := strings.TrimRight(strings.TrimSpace(s.BaseURL), "/")
	if raw == "" {
		return "", fmt.Errorf("target site base URL is required")
	}
	if _, err := url.ParseRequestURI(raw); err != nil {
		return "", fmt.Errorf("target site base URL: %w", err)
	}
	return raw, nil
}
func (s *TargetHTTPService) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: 120 * time.Second}
}
func (s *TargetHTTPService) do(ctx context.Context, client *http.Client, method, endpoint, cookie string, body io.Reader) (*http.Response, error) {
	return s.doWithContentType(ctx, client, method, endpoint, cookie, body, "application/json")
}
func (s *TargetHTTPService) doWithContentType(ctx context.Context, client *http.Client, method, endpoint, cookie string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	return client.Do(req)
}
func mergeCookies(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ";") {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "=") {
				parts = append(parts, part)
				break
			}
		}
	}
	return strings.Join(parts, "; ")
}
