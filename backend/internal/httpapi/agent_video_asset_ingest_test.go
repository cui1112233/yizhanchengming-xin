package httpapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

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
