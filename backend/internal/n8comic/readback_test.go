package n8comic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"
)

func TestCarouselReadbackAfterDisconnectBeforeJSON(t *testing.T) {
	creates, reads := 0, 0
	sigIDs := map[string]int64{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case createPath:
			creates++
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		case readPath:
			reads++
			var body struct {
				Signature string `json:"signature"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			fmt.Fprintf(w, `{"code":"SUCCESS","data":{"image_id":%d,"carousel_id":17}}`, sigIDs[body.Signature])
		default:
			t.Errorf("unexpected %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client := NewClient(Config{BaseURL: server.URL, APIKey: "test", HTTPClient: server.Client()})
	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	ids := []int64{5, 6}
	checked := false
	result, err := client.Publish(context.Background(), PublishRequest{ProducerAdminName: "maker", Name: "c", WorkIDs: []int64{7}, AudioCategoryID: 1, AudioID: 2, Images: []ImageInput{{FileName: "1.png", MediaType: "image/png", Content: png}, {FileName: "2.png", MediaType: "image/png", Content: append(slices.Clone(png), 10)}}, UploadedImageIDs: ids, BeforeCreate: func(proof PublishResult) error {
		if creates != 0 || !slices.Equal(proof.ImageIDs, ids) || len(proof.ImageSignatures) != 2 {
			t.Fatalf("missing pre-create evidence: %+v", proof)
		}
		for i, s := range proof.ImageSignatures {
			sigIDs[s] = ids[i]
		}
		checked = true
		return nil
	}})
	if !errors.Is(err, ErrCarouselReceiptPending) || !checked || creates != 1 || len(result.ImageSignatures) != 2 {
		t.Fatalf("result=%+v checked=%v creates=%d err=%v", result, checked, creates, err)
	}
	got, err := client.Readback(context.Background(), ReceiptLookup{ImageIDs: result.ImageIDs, ImageSignatures: result.ImageSignatures})
	if err != nil || got.CarouselID != 0 || creates != 1 || reads != 0 {
		t.Fatal(got, reads, creates, err)
	}
}

func TestCarouselReadbackDoesNotConfirmMissingOrConflictingOwnership(t *testing.T) {
	for _, parents := range [][]int64{{0, 0}, {9, 9}, {17, 18}} {
		replies := map[string]int64{"a": parents[0], "b": parents[1]}
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != readPath {
				t.Errorf("unexpected %s", r.URL.Path)
			}
			var data struct {
				Signature string `json:"signature"`
			}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Error(err)
				return
			}
			image := map[string]int{"a": 5, "b": 6}[data.Signature]
			fmt.Fprintf(w, `{"code":"SUCCESS","data":{"image_id":%d,"carousel_id":%d}}`, image, replies[data.Signature])
		}))
		c := NewClient(Config{BaseURL: s.URL, APIKey: "test", HTTPClient: s.Client()})
		result, err := c.Readback(context.Background(), ReceiptLookup{ImageIDs: []int64{5, 6}, ImageSignatures: []string{"a", "b"}, PreviousCarouselID: 9})
		s.Close()
		if err != nil || result.CarouselID != 0 {
			t.Fatalf("parents=%v result=%+v err=%v", parents, result, err)
		}
	}
}

func TestCarouselReadbackFollowsOnlyServerProvidedSameOriginLocation(t *testing.T) {
	reads, creates := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == createPath {
			creates++
			w.Header().Set("Location", "/receipts/123")
			w.Header().Set("X-Operation-ID", "provider-op")
			w.WriteHeader(202)
			return
		}
		if r.Method != "GET" || r.URL.Path != "/receipts/123" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			return
		}
		reads++
		if reads == 1 {
			w.WriteHeader(202)
			return
		}
		io.WriteString(w, `{"code":"SUCCESS","data":{"operation_id":"provider-op","carousel_id":41,"image_ids":[5,6]}}`)
	}))
	defer server.Close()
	c := NewClient(Config{BaseURL: server.URL, APIKey: "test", HTTPClient: server.Client()})
	_, err := c.CreateCarousel(context.Background(), CreateCarouselRequest{ProducerAdminName: "m", Name: "book", ImageIDs: []int64{5, 6}, WorkIDs: []int64{7}, AudioCategoryID: 1, AudioID: 2})
	var pending *pendingReceipt
	if !errors.As(err, &pending) || pending.path != "/receipts/123" {
		t.Fatal(err)
	}
	for _, value := range []string{"https://example.invalid/private", "//example.invalid/receipt", "http://user@example.invalid/r"} {
		if c.receiptPath(value) != "" {
			t.Fatal(value)
		}
	}
	lookup := ReceiptLookup{IntentID: "local-attempt", OperationID: pending.operation, ReceiptPath: pending.path, ImageIDs: []int64{5, 6}}
	first, err := c.Readback(context.Background(), lookup)
	if err != nil || first.CarouselID != 0 {
		t.Fatal(first, err)
	}
	next, err := c.Readback(context.Background(), lookup)
	if err != nil || next.CarouselID != 41 || reads != 2 || creates != 1 {
		t.Fatal(next, err, reads, creates)
	}
}

func TestCarouselDefiniteRejectionAndAmbiguousResponse(t *testing.T) {
	for _, tc := range []struct {
		status  int
		body    string
		pending bool
	}{{400, "", false}, {200, `{"code":"INVALID_AUDIO"}`, false}, {502, "", true}, {200, `{"code":"SUCCESS",`, true}, {200, `{}`, true}} {
		t.Run(fmt.Sprintf("%d-%s", tc.status, tc.body), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); io.WriteString(w, tc.body) }))
			defer s.Close()
			c := NewClient(Config{BaseURL: s.URL, APIKey: "test", HTTPClient: s.Client()})
			_, err := c.CreateCarousel(context.Background(), CreateCarouselRequest{ProducerAdminName: "m", Name: "book", ImageIDs: []int64{5, 6}, WorkIDs: []int64{7}, AudioCategoryID: 1, AudioID: 2})
			if err == nil || errors.Is(err, ErrCarouselReceiptPending) != tc.pending {
				t.Fatal(err)
			}
		})
	}
}

func TestCarouselReceiptRequiresMatchingOperationAndExactPageOrder(t *testing.T) {
	for _, tc := range []struct {
		name, body    string
		wantID        int64
		wantRejection bool
	}{
		{"missing identity", `{"code":"SUCCESS","data":{"image_ids":[1,2],"carousel_id":7}}`, 0, false},
		{"other operation", `{"code":"SUCCESS","data":{"request_id":"other","image_ids":[1,2],"carousel_id":7}}`, 0, false},
		{"wrong order", `{"code":"SUCCESS","data":{"request_id":"here","image_ids":[2,1],"carousel_id":7}}`, 0, false},
		{"missing target", `{"code":"SUCCESS","data":{"request_id":"here","carousel_id":7}}`, 0, false},
		{"running", `{"code":"SUCCESS","data":{"status":"PROCESSING","request_id":"here","image_ids":[1,2],"carousel_id":7}}`, 0, false},
		{"unrelated rejection", `{"code":"SUCCESS","data":{"status":"REJECTED","request_id":"other","reason":"no"}}`, 0, false},
		{"rejection", `{"code":"SUCCESS","data":{"status":"REJECTED","request_id":"here","reason":"拒绝原因"}}`, 0, true},
		{"matching", `{"code":"SUCCESS","data":{"request_id":"here","image_ids":[1,2],"carousel_id":7}}`, 7, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/status/opaque" {
					t.Fatal("unexpected read")
				}
				io.WriteString(w, tc.body)
			}))
			defer s.Close()
			c := NewClient(Config{BaseURL: s.URL, APIKey: "test", HTTPClient: s.Client()})
			result, err := c.Readback(context.Background(), ReceiptLookup{IntentID: "here", ReceiptPath: "/status/opaque", ImageIDs: []int64{1, 2}})
			var rejected *RejectedReceipt
			if errors.As(err, &rejected) != tc.wantRejection || result.CarouselID != tc.wantID || rejected != nil && (rejected.IntentID != "here" || rejected.Reason != "拒绝原因") {
				t.Fatal(result, err)
			}
		})
	}
}

func TestCreateCarouselRedirectIsNeverFollowed(t *testing.T) {
	var foreignCalls atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { foreignCalls.Add(1); w.WriteHeader(404) }))
	defer other.Close()
	for _, code := range []int{301, 302, 303, 307, 308} {
		for _, foreign := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/foreign=%v", code, foreign), func(t *testing.T) {
				var creates, reads atomic.Int32
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != createPath {
						reads.Add(1)
						w.WriteHeader(404)
						return
					}
					creates.Add(1)
					location := "/receipts/op"
					if foreign {
						location = other.URL + location
					}
					w.Header().Set("Location", location)
					w.Header().Set("X-Operation-ID", "op")
					w.WriteHeader(code)
				}))
				defer s.Close()
				c := NewClient(Config{BaseURL: s.URL, APIKey: "local", HTTPClient: s.Client()})
				_, err := c.CreateCarousel(context.Background(), CreateCarouselRequest{ProducerAdminName: "p", Name: "book", ImageIDs: []int64{3, 4}, WorkIDs: []int64{5}, AudioCategoryID: 1, AudioID: 2})
				var pending *pendingReceipt
				if !errors.As(err, &pending) || !errors.Is(err, ErrCarouselReceiptPending) || creates.Load() != 1 || reads.Load() != 0 || foreignCalls.Load() != 0 {
					t.Fatalf("err=%v creates=%d reads=%d foreign=%d", err, creates.Load(), reads.Load(), foreignCalls.Load())
				}
				want := "/receipts/op"
				if foreign {
					want = ""
				}
				if pending.path != want || pending.operation != "op" {
					t.Fatalf("pending=%+v", pending)
				}
			})
		}
	}
}

func TestReadbackUsesOperationIdentityIndependentlyOfQueryTrace(t *testing.T) {
	for _, state := range []string{"SUCCESS", "REJECTED", "trace-only"} {
		t.Run(state, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("unexpected method %s", r.Method)
				}
				w.Header().Set("X-Request-ID", "new-query-trace")
				if state == "trace-only" {
					w.Header().Set("X-Request-ID", "create-intent")
					io.WriteString(w, `{"code":"SUCCESS","data":{"image_ids":[3,4],"carousel_id":9}}`)
					return
				}
				fmt.Fprintf(w, `{"code":"SUCCESS","data":{"operation_id":"op","status":%q,"image_ids":[3,4],"carousel_id":9,"reason":"publisher rejected"}}`, state)
			}))
			defer s.Close()
			c := NewClient(Config{BaseURL: s.URL, APIKey: "local", HTTPClient: s.Client()})
			result, err := c.Readback(context.Background(), ReceiptLookup{ReceiptPath: "/receipt", IntentID: "create-intent", OperationID: "op", ImageIDs: []int64{3, 4}})
			var rejected *RejectedReceipt
			switch state {
			case "SUCCESS":
				if err != nil || result.CarouselID != 9 || result.ReceiptIntentID != "create-intent" {
					t.Fatal(result, err)
				}
			case "REJECTED":
				if !errors.As(err, &rejected) || rejected.IntentID != "create-intent" || rejected.Reason != "publisher rejected" {
					t.Fatal(result, err)
				}
			default:
				if err != nil || result.CarouselID != 0 {
					t.Fatal(result, err)
				}
			}
		})
	}
}
