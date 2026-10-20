package httpapi

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func TestAgentMediaAssetIngestSavesPNGDataURL(t *testing.T) {
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &recordingAgentNodeObjectStore{}
	ingestor := newAgentMediaAssetIngestor(assets, objects, nil, func() time.Time {
		return time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	})
	asset, err := ingestor.SaveImageResult(context.Background(), "alice", "project-1", "asset-1", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Kind != "image" || asset.ContentType != "image/png" || asset.URL != "/api/agent/tasks/project-1/assets/asset-1/content" {
		t.Fatalf("asset=%+v", asset)
	}
	if len(objects.putContents) == 0 {
		t.Fatal("image bytes were not stored")
	}
}

func TestAgentMediaAssetIngestRejectsNonImageDataURL(t *testing.T) {
	ingestor := newAgentMediaAssetIngestor(agentworkspace.NewMemoryAssetStore(), &recordingAgentNodeObjectStore{}, nil, time.Now)
	if _, err := ingestor.SaveImageResult(context.Background(), "alice", "project-1", "asset-1", "data:text/plain;base64,aGVsbG8="); err == nil {
		t.Fatal("non-image data URL was accepted")
	}
}

func TestAgentMediaAssetIngestRejectsBodyThatLiesAboutImageType(t *testing.T) {
	ingestor := newAgentMediaAssetIngestor(agentworkspace.NewMemoryAssetStore(), &recordingAgentNodeObjectStore{}, nil, time.Now)
	ingestor.fetch = func(context.Context, string) (io.ReadCloser, string, error) {
		return io.NopCloser(strings.NewReader("this is not an image")), "image/png", nil
	}
	if _, err := ingestor.SaveImageResult(context.Background(), "alice", "project-1", "asset-1", "https://assets.example/result.png"); err == nil {
		t.Fatal("non-image body with image content type was accepted")
	}
}

func TestAgentMediaAssetIngestRollsBackObjectWhenAssetCreateFails(t *testing.T) {
	objects := &recordingAgentNodeObjectStore{}
	ingestor := newAgentMediaAssetIngestor(failingAgentNodeAssetStore{err: errors.New("database unavailable")}, objects, nil, time.Now)
	_, err := ingestor.SaveImageResult(context.Background(), "alice", "project-1", "asset-1", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB")
	if err == nil {
		t.Fatal("asset store failure was ignored")
	}
	if len(objects.deleted) != 1 || objects.deleted[0] != "asset-1" {
		t.Fatalf("deleted=%v", objects.deleted)
	}
}

func TestAgentVideoSafeDialRejectsPrivateResolvedAddressBeforeConnect(t *testing.T) {
	dialCalls := 0
	dial := newAgentVideoSafeDialContext(
		func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
		},
		func(context.Context, string, string) (net.Conn, error) {
			dialCalls++
			return nil, nil
		},
	)
	if _, err := dial(context.Background(), "tcp", "provider.example:443"); err == nil || dialCalls != 0 {
		t.Fatalf("err=%v dial calls=%d, want private address rejected before connect", err, dialCalls)
	}
}

func TestAgentVideoRemoteFetchRejectsPrivateDestinationBeforeTransport(t *testing.T) {
	transportCalls := 0
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		transportCalls++
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"video/mp4"}}, Body: io.NopCloser(strings.NewReader("mp4"))}, nil
	})}
	ingestor := newAgentVideoAssetIngestor(nil, nil, client, nil)

	_, _, err := ingestor.fetch(context.Background(), "https://127.0.0.1/private.mp4")
	if err == nil || transportCalls != 0 {
		t.Fatalf("err=%v transport calls=%d, want rejection before transport", err, transportCalls)
	}
}

func TestAgentVideoRemoteFetchRejectsUnsafeRedirectBeforeSecondRequest(t *testing.T) {
	for _, location := range []string{"http://93.184.216.34/private.mp4", "https://127.0.0.1/private.mp4", "https://169.254.169.254/latest/meta-data"} {
		t.Run(location, func(t *testing.T) {
			transportCalls := 0
			client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				transportCalls++
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": []string{location}},
					Body:       io.NopCloser(strings.NewReader("")),
					Request:    req,
				}, nil
			})}
			ingestor := newAgentVideoAssetIngestor(nil, nil, client, nil)

			_, _, err := ingestor.fetch(context.Background(), "https://93.184.216.34/video.mp4")
			if err == nil {
				t.Fatal("unsafe redirect was accepted")
			}
			if transportCalls != 1 {
				t.Fatalf("transport calls=%d, unsafe redirect request was sent", transportCalls)
			}
		})
	}
}
