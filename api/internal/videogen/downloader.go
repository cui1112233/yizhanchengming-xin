package videogen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultMaxVideoDownloadBytes int64 = 1024 << 20 // 1 GiB

type HTTPDownloader struct {
	Client      *http.Client
	ValidateURL URLValidator
	MaxBytes    int64
}

func (d HTTPDownloader) Download(ctx context.Context, raw string) (DownloadedMedia, error) {
	raw = strings.TrimSpace(raw)
	if err := d.validate(raw); err != nil {
		return DownloadedMedia{}, fmt.Errorf("unsafe generated video URL: %w", err)
	}
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil { return DownloadedMedia{}, err }
	resp, err := client.Do(req)
	if err != nil { return DownloadedMedia{}, fmt.Errorf("download generated video: %w", err) }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return DownloadedMedia{}, fmt.Errorf("generated video download returned HTTP %d", resp.StatusCode)
	}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	if !strings.HasPrefix(contentType, "video/") {
		resp.Body.Close()
		return DownloadedMedia{}, fmt.Errorf("generated video response has invalid content type %q", contentType)
	}
	maxBytes := d.MaxBytes
	if maxBytes <= 0 { maxBytes = defaultMaxVideoDownloadBytes }
	if resp.ContentLength > maxBytes {
		resp.Body.Close()
		return DownloadedMedia{}, fmt.Errorf("generated video exceeds %d bytes", maxBytes)
	}
	body := &limitedReadCloser{Reader: io.LimitReader(resp.Body, maxBytes+1), Closer: resp.Body, max: maxBytes}
	return DownloadedMedia{Body: body, ContentType: contentType, ContentLength: resp.ContentLength}, nil
}

func (d HTTPDownloader) validate(raw string) error {
	if d.ValidateURL != nil { return d.ValidateURL(raw) }
	u, err := url.Parse(raw)
	if err != nil { return err }
	if u.Scheme != "https" && u.Scheme != "http" { return errors.New("media URL must use http or https") }
	host := strings.TrimSpace(u.Hostname())
	if host == "" { return errors.New("media URL host is required") }
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") { return errors.New("local host is forbidden") }
	if ip := net.ParseIP(host); ip != nil && !publicIP(ip) { return errors.New("private or local IP is forbidden") }
	return nil
}

type limitedReadCloser struct {
	Reader io.Reader
	Closer io.Closer
	max int64
	read int64
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.read += int64(n)
	if r.read > r.max { return n, errors.New("generated video exceeds download size limit") }
	return n, err
}
func (r *limitedReadCloser) Close() error { return r.Closer.Close() }

func publicIP(ip net.IP) bool {
	if ip == nil { return false }
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() { return false }
	return true
}

func parseContentLength(value string) int64 {
	result, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return result
}
