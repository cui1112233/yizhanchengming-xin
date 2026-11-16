package subjectupload

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRejectMissingConfigBeforeNetwork(t *testing.T) {
	c := NewClient()
	_, err := c.Run(context.Background(), Config{Provider: "volcano"}, Input{Name: "角色", ImageURL: "https://example.com/a.png"}, Progress{}, nil)
	var re *RunError
	if !errors.As(err, &re) || re.Kind != "failed" {
		t.Fatalf("expected safe config error: %v", err)
	}
}

func TestOriginalSignatureVector(t *testing.T) {
	body := []byte(`{"Name":"OWN_SUBJECT","Description":"Character asset group for OWN_SUBJECT","GroupType":"AIGC","ProjectName":"OWN_PROJECT"}`)
	req, _ := http.NewRequest("POST", "https://ark.cn-beijing.volcengineapi.com/?Action=CreateAssetGroup&Version=2024-01-01", nil)
	now, _ := time.Parse("20060102T150405Z", "20261115T222730Z")
	signFire(req, body, Config{AccessKeyID: "OWN_FAKE_AK", SecretAccessKey: "OWN_FAKE_SK"}, now)
	if !strings.HasSuffix(req.Header.Get("Authorization"), "Signature=af00af8b97c29ea7d9ab316a29344e18c78181c7a17a46ec1f7cee8e00db2804") {
		t.Fatal(req.Header.Get("Authorization"))
	}
}

func TestFireOriginalRequestsAndCheckpoints(t *testing.T) {
	var p Progress
	var actions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.URL.Query().Get("Action")
		actions = append(actions, a)
		if a == "CreateAssetGroup" && p.PendingAction != "create_group" {
			t.Error("missing pre-create checkpoint")
		}
		if a == "CreateAsset" && (p.PendingAction != "create_asset" || p.GroupID != "OWN_GROUP") {
			t.Error("missing pre-asset checkpoint")
		}
		b, _ := io.ReadAll(r.Body)
		switch a {
		case "CreateAssetGroup":
			if string(b) != `{"Name":"OWN_SUBJECT","Description":"Character asset group for OWN_SUBJECT","GroupType":"AIGC","ProjectName":"OWN_PROJECT"}` {
				t.Error(string(b))
			}
			io.WriteString(w, `{"Result":{"Id":"OWN_GROUP"}}`)
		case "CreateAsset":
			if string(b) != `{"GroupId":"OWN_GROUP","URL":"https://own.invalid/image.png","AssetType":"Image","ProjectName":"OWN_PROJECT","Name":"own.png"}` {
				t.Error(string(b))
			}
			io.WriteString(w, `{"Result":{"Id":"OWN_ASSET"}}`)
		case "GetAsset":
			io.WriteString(w, `{"Result":{"Id":"OWN_ASSET","Status":"Active"}}`)
		}
	}))
	defer server.Close()
	c := NewClient()
	c.HTTP = server.Client()
	c.FireEndpoint = server.URL
	res, e := c.Run(context.Background(), Config{Provider: "volcano", AccessKeyID: "OWN_FAKE_AK", SecretAccessKey: "OWN_FAKE_SK", ProjectName: "OWN_PROJECT"}, Input{Name: "OWN_SUBJECT", FileName: "own.png", ImageURL: "https://own.invalid/image.png"}, Progress{}, func(_ context.Context, next Progress) error { p = next; return nil })
	if e != nil || res.AssetID != "OWN_ASSET" || p.PendingAction != "" || len(actions) != 3 {
		t.Fatalf("%+v %v %+v", res, e, p)
	}
}

func TestDeyunWaitCancelAndResumeOnlyQueries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" {
			t.Error("resume must only query")
		}
		cancel()
		json.NewEncoder(w).Encode(map[string]string{"Id": "OWN_ASSET", "Status": "Processing"})
	}))
	defer s.Close()
	c := NewClient()
	c.HTTP = s.Client()
	c.DeyunaiEndpoint = s.URL
	c.PollInterval = time.Hour
	_, e := c.Run(ctx, Config{Provider: "deyunai", Region: "c39", Token: "OWN_TOKEN"}, Input{Name: "own"}, Progress{AssetID: "OWN_ASSET"}, nil)
	var re *RunError
	if !errors.As(e, &re) || re.Kind != "paused" || calls != 1 || re.Progress.AssetID != "OWN_ASSET" {
		t.Fatalf("%v %d", e, calls)
	}
}

func TestDeyunOriginalProcessingProtocol(t *testing.T) {
	for _, region := range []string{"c39", "c40"} {
		t.Run(region, func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer OWN_GLOBAL_TOKEN" || r.Header.Get("Accept") != "application/json" {
					t.Error("headers differ")
				}
				if calls == 1 {
					b, _ := io.ReadAll(r.Body)
					if r.Method != "POST" || r.URL.Path != "/"+region+"/v1/sd/assets" || string(b) != `{"URL":"https://own.invalid/image.png","Name":"OWN_SUBJECT","AssetType":"Image"}` {
						t.Fatalf("request mismatch %s %s %s", r.Method, r.URL.Path, b)
					}
					io.WriteString(w, `{"Id":"OWN_ASSET","Status":"Processing"}`)
				} else {
					if r.Method != "GET" || r.URL.Path != "/"+region+"/v1/sd/assets/OWN_ASSET" || r.Header.Get("Content-Type") != "" {
						t.Error("query mismatch")
					}
					io.WriteString(w, `{"Id":"OWN_ASSET","Status":"Active"}`)
				}
			}))
			defer s.Close()
			c := NewClient()
			c.HTTP = s.Client()
			c.DeyunaiEndpoint = s.URL
			c.PollInterval = time.Millisecond
			got, e := c.Run(context.Background(), Config{Provider: "deyunai", Token: "OWN_GLOBAL_TOKEN", Region: region}, Input{Name: "OWN_SUBJECT", Description: "ignored as original", ImageURL: "https://own.invalid/image.png"}, Progress{}, nil)
			if e != nil || calls != 2 || got.ProjectName != "deyunai" || got.Status != "Active" {
				t.Fatalf("%+v %v", got, e)
			}
		})
	}
}
func TestPendingCreationIsNeverRepeated(t *testing.T) {
	_, err := NewClient().Run(context.Background(), Config{}, Input{}, Progress{PendingAction: "create_asset"}, nil)
	var re *RunError
	if !errors.As(err, &re) || re.Kind != "needs_review" || re.Progress.PendingAction != "create_asset" {
		t.Fatalf("lost uncertain progress: %v", err)
	}
}

func TestProviderFailureBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, create, get, kind string
		wantCalls               int
	}{
		{"no_id", `{}`, ``, "needs_review", 1},
		{"create_error", `{"error":"OWN_FAKE_TOKEN"}`, ``, "needs_review", 1},
		{"oversize", strings.Repeat("x", (1<<20)+1), ``, "needs_review", 1},
		{"failed", `{"Id":"OWN_ASSET","Status":"Processing"}`, `{"Id":"OWN_ASSET","Status":"Failed"}`, "failed", 2},
		{"unknown", `{"Id":"OWN_ASSET","Status":"Unrecognized"}`, ``, "paused", 1},
		{"waiting", `{"Id":"OWN_ASSET","Status":"Processing"}`, `{"Id":"OWN_ASSET","Status":"Processing"}`, "paused", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method == "POST" {
					io.WriteString(w, tc.create)
				} else {
					io.WriteString(w, tc.get)
				}
			}))
			defer s.Close()
			c := NewClient()
			c.HTTP = s.Client()
			c.DeyunaiEndpoint = s.URL
			c.MaxWait = 30 * time.Millisecond
			c.PollInterval = 20 * time.Millisecond
			var saved Progress
			_, e := c.Run(context.Background(), Config{Provider: "deyunai", Region: "c40", Token: "OWN_FAKE_TOKEN"}, Input{Name: "OWN_SUBJECT", ImageURL: "https://own.invalid/a.png"}, Progress{}, func(_ context.Context, p Progress) error { saved = p; return nil })
			var re *RunError
			if !errors.As(e, &re) || re.Kind != tc.kind {
				t.Fatalf("%v", e)
			}
			if strings.Contains(e.Error(), "OWN_FAKE_TOKEN") {
				t.Error("credential leak")
			}
			if calls != tc.wantCalls {
				t.Errorf("calls=%d", calls)
			}
			if tc.kind == "needs_review" && saved.PendingAction != "create_asset" {
				t.Fatal("unknown creation lost")
			}
		})
	}
}

func TestImageHostMultipartAndURLValidation(t *testing.T) {
	for _, target := range []string{"https://own.invalid/safe.png", "http://own.invalid/no.png", "https://127.0.0.1/no.png", "https://user:password@own.invalid/no.png"} {
		t.Run(target, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer OWN_HOST_TOKEN" {
					t.Error("missing host token")
				}
				file, h, e := r.FormFile("file")
				if e != nil {
					t.Fatal(e)
				}
				defer file.Close()
				b, _ := io.ReadAll(file)
				if string(b) != "image" || h.Filename != "own.png" {
					t.Error("wrong multipart")
				}
				json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"url": target}})
			}))
			defer s.Close()
			c := NewClient()
			c.HTTP = s.Client()
			u, e := c.hostImage(context.Background(), Config{ImageHostURL: s.URL, ImageHostToken: "OWN_HOST_TOKEN"}, Input{Image: []byte("image"), FileName: "own.png", ContentType: "image/png"})
			if target == "https://own.invalid/safe.png" {
				if e != nil || u != target {
					t.Fatal(e)
				}
			} else if e == nil {
				t.Fatal("unsafe URL accepted")
			}
		})
	}
}

func TestCheckpointFailurePreventsCreate(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer s.Close()
	c := NewClient()
	c.HTTP = s.Client()
	c.DeyunaiEndpoint = s.URL
	_, e := c.Run(context.Background(), Config{Provider: "deyunai", Region: "c40", Token: "OWN"}, Input{Name: "own"}, Progress{ImageURL: "https://own.invalid/a.png"}, func(context.Context, Progress) error { return errors.New("storage down") })
	if e == nil || calls != 0 {
		t.Fatalf("%v %d", e, calls)
	}
}

func TestRedirectDoesNotForwardCredentials(t *testing.T) {
	reached := false
	dest := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))
	defer dest.Close()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, dest.URL, http.StatusTemporaryRedirect)
	}))
	defer s.Close()
	c := NewClient()
	c.HTTP = s.Client()
	_, e := c.request(context.Background(), "POST", s.URL, nil, map[string]string{"Authorization": "Bearer OWN"})
	if e == nil || reached {
		t.Fatal("followed redirect")
	}
}

func TestImageHostFailureCanResumeSameProgress(t *testing.T) {
	hostCalls, createCalls := 0, 0
	hostFixed := false
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostCalls++
		if !hostFixed {
			http.Error(w, "temporary host failure", http.StatusServiceUnavailable)
			return
		}
		io.WriteString(w, `{"url":"https://own.invalid/image.png"}`)
	}))
	defer host.Close()
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("unexpected provider query")
		}
		createCalls++
		io.WriteString(w, `{"Id":"OWN_ASSET","Status":"Active"}`)
	}))
	defer provider.Close()
	c := NewClient()
	c.HTTP = host.Client()
	hostTarget, _ := url.Parse(host.URL)
	baseTransport := c.HTTP.Transport
	c.HTTP.Transport = uploadRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "image-host.invalid" {
			clone := r.Clone(r.Context())
			copyURL := *r.URL
			copyURL.Host = hostTarget.Host
			clone.URL = &copyURL
			return baseTransport.RoundTrip(clone)
		}
		return baseTransport.RoundTrip(r)
	})
	c.DeyunaiEndpoint = provider.URL
	cfg := Config{Provider: "deyunai", Region: "c40", Token: "OWN_TOKEN", ImageHostURL: "https://image-host.invalid"}
	in := Input{Name: "own", FileName: "own.png", ContentType: "image/png", Image: []byte("image")}
	_, err := c.Run(context.Background(), cfg, in, Progress{}, nil)
	var stopped *RunError
	if !errors.As(err, &stopped) || stopped.Kind != "paused" || stopped.Progress.PendingAction != "" || createCalls != 0 {
		t.Fatalf("unexpected first attempt: %v, creates=%d", err, createCalls)
	}
	hostFixed = true
	got, err := c.Run(context.Background(), cfg, in, stopped.Progress, nil)
	if err != nil || got.AssetID != "OWN_ASSET" || hostCalls != 2 || createCalls != 1 {
		t.Fatalf("resume result=%+v err=%v hosts=%d creates=%d", got, err, hostCalls, createCalls)
	}
}

type uploadRoundTripFunc func(*http.Request) (*http.Response, error)

func (f uploadRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
