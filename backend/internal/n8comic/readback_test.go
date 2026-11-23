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
	if err != nil || got.CarouselID != 17 || !slices.Equal(got.ImageIDs, ids) || creates != 1 || reads != 2 {
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
		io.WriteString(w, `{"code":"SUCCESS","data":{"carousel_id":41,"image_ids":[5,6]}}`)
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
	lookup := ReceiptLookup{ReceiptPath: pending.path, ImageIDs: []int64{5, 6}}
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
