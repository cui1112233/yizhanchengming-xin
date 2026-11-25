package n8comic

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type parallelUploadServer struct {
	server     *httptest.Server
	started    chan struct{}
	release    chan struct{}
	once       sync.Once
	active     atomic.Int32
	peak       atomic.Int32
	creates    atomic.Int32
	puts       atomic.Int32
	signatures map[string]int64
}

func newParallelUploadServer(t *testing.T, count int) (*parallelUploadServer, []ImageInput) {
	t.Helper()
	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	s := &parallelUploadServer{started: make(chan struct{}, 64), release: make(chan struct{}), signatures: map[string]int64{}}
	images := make([]ImageInput, count)
	for i := range images {
		content := append(append([]byte(nil), png...), byte(i))
		digest := md5.Sum(content)
		s.signatures[hex.EncodeToString(digest[:])] = int64(700 + i)
		images[i] = ImageInput{FileName: fmt.Sprintf("page-%02d.png", i), MediaType: "image/png", Content: content}
	}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
		}
		switch r.URL.Path {
		case uploadPath:
			sig, _ := body["signature"].(string)
			id := s.signatures[sig]
			if id == 0 {
				t.Error("unknown image signature")
				w.WriteHeader(400)
				return
			}
			fmt.Fprintf(w, `{"code":"SUCCESS","data":{"url":%q}}`, s.server.URL+fmt.Sprintf("/put/%d", id))
		case readPath:
			sig, _ := body["signature"].(string)
			fmt.Fprintf(w, `{"code":"SUCCESS","data":{"image_id":%d}}`, s.signatures[sig])
		case createPath:
			s.creates.Add(1)
			fmt.Fprint(w, `{"code":"SUCCESS","data":{"carousel_id":1234}}`)
		default:
			if r.Method != http.MethodPut {
				w.WriteHeader(404)
				return
			}
			s.puts.Add(1)
			n := s.active.Add(1)
			defer s.active.Add(-1)
			for old := s.peak.Load(); n > old; old = s.peak.Load() {
				if s.peak.CompareAndSwap(old, n) {
					break
				}
			}
			s.started <- struct{}{}
			select {
			case <-s.release:
				w.WriteHeader(200)
			case <-r.Context().Done():
			}
		}
	}))
	t.Cleanup(func() { s.unblock(); s.server.Close() })
	return s, images
}

func (s *parallelUploadServer) unblock() { s.once.Do(func() { close(s.release) }) }

func awaitUploadStarts(s *parallelUploadServer, count int) int {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for n := 0; n < count; n++ {
		select {
		case <-s.started:
		case <-timer.C:
			return n
		}
	}
	return count
}

func TestPublishTransfersPagesInParallelAndKeepsOrder(t *testing.T) {
	s, images := newParallelUploadServer(t, 7)
	c := NewClient(Config{BaseURL: s.server.URL, APIKey: "test", HTTPClient: s.server.Client()})
	type outcome struct {
		result PublishResult
		err    error
	}
	done := make(chan outcome, 1)
	callbacks := make([]int64, len(images))
	go func() {
		r, err := c.Publish(context.Background(), PublishRequest{ProducerAdminName: "maker", Name: "parallel", WorkIDs: []int64{8}, AudioCategoryID: 1, AudioID: 2, Images: images, OnImage: func(i int, id int64) error { callbacks[i] = id; return nil }})
		done <- outcome{r, err}
	}()
	started := awaitUploadStarts(s, 3)
	s.unblock()
	o := <-done
	if started != 3 {
		t.Fatalf("only %d transfers started before release; want 3 concurrent transfers", started)
	}
	want := []int64{700, 701, 702, 703, 704, 705, 706}
	if o.err != nil || !reflect.DeepEqual(o.result.ImageIDs, want) || !reflect.DeepEqual(callbacks, want) || s.creates.Load() != 1 {
		t.Fatalf("result=%+v callbacks=%v creates=%d err=%v", o.result, callbacks, s.creates.Load(), o.err)
	}
	if s.peak.Load() > 3 {
		t.Fatalf("per-collection transfers=%d, want at most 3", s.peak.Load())
	}
}

func TestParallelImageTransfersShareClientLimit(t *testing.T) {
	s, images := newParallelUploadServer(t, 9)
	c := NewClient(Config{BaseURL: s.server.URL, APIKey: "test", HTTPClient: s.server.Client()})
	var wg sync.WaitGroup
	errors := make(chan error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := c.UploadImages(context.Background(), UploadImagesRequest{ProducerAdminName: "maker", Images: images[i*3 : (i+1)*3]})
			errors <- err
		}(i)
	}
	started := awaitUploadStarts(s, 6)
	s.unblock()
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
	if started != 6 || s.peak.Load() > 6 {
		t.Fatalf("started=%d peak=%d, want 6 shared slots", started, s.peak.Load())
	}
	if s.creates.Load() != 0 {
		t.Fatal("image transfer created a carousel")
	}
}

func TestParallelImageTransferResumeSkipsKnownPages(t *testing.T) {
	s, images := newParallelUploadServer(t, 4)
	s.unblock()
	c := NewClient(Config{BaseURL: s.server.URL, APIKey: "test", HTTPClient: s.server.Client()})
	ids, err := c.UploadImages(context.Background(), UploadImagesRequest{ProducerAdminName: "maker", Images: images, UploadedImageIDs: []int64{700, 0, 702, 0}})
	if err != nil || !reflect.DeepEqual(ids, []int64{700, 701, 702, 703}) || s.puts.Load() != 2 {
		t.Fatalf("ids=%v transfers=%d err=%v", ids, s.puts.Load(), err)
	}
}

func TestParallelTransfersCancelWithoutCreatingCarousel(t *testing.T) {
	s, images := newParallelUploadServer(t, 7)
	c := NewClient(Config{BaseURL: s.server.URL, APIKey: "test", HTTPClient: s.server.Client()})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := c.Publish(ctx, PublishRequest{ProducerAdminName: "maker", Name: "cancel", WorkIDs: []int64{8}, AudioCategoryID: 1, AudioID: 2, Images: images})
		done <- err
	}()
	started := awaitUploadStarts(s, 3)
	cancel()
	s.unblock()
	select {
	case err := <-done:
		if err == nil || s.creates.Load() != 0 || started != 3 {
			t.Fatalf("err=%v creates=%d started=%d", err, s.creates.Load(), started)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled transfer did not finish")
	}
}
