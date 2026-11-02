package n8comic

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClientPublishesOrderedImagesAndCreatesCarousel(t *testing.T) {
	now := time.Unix(1700000000, 0)
	wantSignBytes := md5.Sum([]byte("server-secret" + strconv.FormatInt(now.Unix(), 10)))
	wantSign := hex.EncodeToString(wantSignBytes[:])
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}

	var server *httptest.Server
	readAttempts := map[string]int{}
	uploaded := []string{}
	events := []string{}
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode %s: %v", r.URL.Path, err)
				return
			}
			if got := body["sign"]; got != wantSign {
				t.Errorf("%s sign=%v want=%s", r.URL.Path, got, wantSign)
			}
			if got := int64(body["time"].(float64)); got != now.Unix() {
				t.Errorf("%s time=%d want=%d", r.URL.Path, got, now.Unix())
			}
		}
		switch r.URL.Path {
		case "/material/open/carousel/image_upload_param":
			if body["producer_admin_name"] != "崔家铭" || body["file_width"] != float64(1) || body["file_height"] != float64(1) || body["mime_type"] != "image/png" {
				t.Errorf("upload params=%#v", body)
			}
			signature := body["signature"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{
				"url":    server.URL + "/upload/" + signature,
				"header": map[string]string{"x-tos-callback": "callback", "x-tos-callback-var": "callback-var"},
			}})
		case "/upload/" + bodyStringFromPath(r.URL.Path):
			t.Fatal("unexpected dynamic upload switch branch")
		case "/material/open/carousel/image/read":
			events = append(events, "read")
			signature := body["signature"].(string)
			readAttempts[signature]++
			if readAttempts[signature] == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "IMAGE_NOT_READY", "message": "图片尚未上传完成"})
				return
			}
			id := 101
			if len(readAttempts) > 1 {
				id = 102
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"image_id": id, "signature": signature}})
		case "/material/open/carousel/create":
			if body["producer_admin_name"] != "崔家铭" || body["name"] != "崔家铭1026-测试书" || body["audio_category_id"] != float64(3) || body["audio_id"] != float64(77) {
				t.Errorf("create body=%#v", body)
			}
			if got := toIntSlice(body["image_ids"]); !reflect.DeepEqual(got, []int64{101, 102}) {
				t.Errorf("image_ids=%v", got)
			}
			if got := toIntSlice(body["work_ids"]); !reflect.DeepEqual(got, []int64{2001, 2002}) {
				t.Errorf("work_ids=%v", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"carousel_id": 456}})
		default:
			if r.Method == http.MethodPut && len(r.URL.Path) > len("/upload/") && r.URL.Path[:len("/upload/")] == "/upload/" {
				events = append(events, "put")
				if r.Header.Get("x-tos-callback") != "callback" || r.Header.Get("x-tos-callback-var") != "callback-var" {
					t.Errorf("upload headers=%v", r.Header)
				}
				uploaded = append(uploaded, r.URL.Path)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIKey: "server-secret", HTTPClient: server.Client(), Now: func() time.Time { return now }, PollInterval: time.Millisecond, PollAttempts: 2})
	secondPNG := append(append([]byte(nil), pngBytes...), '\n')
	result, err := client.Publish(context.Background(), PublishRequest{
		ProducerAdminName: "崔家铭",
		Name:              "崔家铭1026-测试书",
		WorkIDs:           []int64{2001, 2002},
		AudioCategoryID:   3,
		AudioID:           77,
		Images: []ImageInput{
			{FileName: "测试书-01.png", MediaType: "image/png", Content: pngBytes},
			{FileName: "测试书-02.png", MediaType: "image/png", Content: secondPNG},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CarouselID != 456 || !reflect.DeepEqual(result.ImageIDs, []int64{101, 102}) || len(uploaded) != 2 {
		t.Fatalf("result=%+v uploaded=%v", result, uploaded)
	}
	if !reflect.DeepEqual(events, []string{"put", "put", "read", "read", "read", "read"}) {
		t.Fatalf("N8 upload/read order=%v", events)
	}
}

func TestClientRejectsIncompletePublishBeforeNetwork(t *testing.T) {
	client := NewClient(Config{BaseURL: "https://n8.example", APIKey: "secret"})
	_, err := client.Publish(context.Background(), PublishRequest{ProducerAdminName: "崔家铭", Name: "合集", WorkIDs: []int64{1}, AudioCategoryID: 3, AudioID: 77, Images: []ImageInput{{FileName: "only.png", MediaType: "image/png", Content: []byte("bad")}}})
	if err == nil || err.Error() != "漫画合集必须包含 2 到 35 张图片" {
		t.Fatalf("err=%v", err)
	}
}

func TestAPIErrorIncludesResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "PRODUCER_ADMIN_ORGANIZATION_INVALID", "message": "制作者不属于指定组织", "data": map[string]any{"hint": "check organization"}})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	err := client.post(context.Background(), "/material/open/carousel/image_upload_param", map[string]any{"producer_admin_name": "崔家铭2"}, nil)
	if err == nil || !strings.Contains(err.Error(), `"code":"PRODUCER_ADMIN_ORGANIZATION_INVALID"`) || !strings.Contains(err.Error(), "响应JSON") {
		t.Fatalf("err=%v", err)
	}
}

func toIntSlice(value any) []int64 {
	values, _ := value.([]any)
	result := make([]int64, 0, len(values))
	for _, item := range values {
		result = append(result, int64(item.(float64)))
	}
	return result
}

func bodyStringFromPath(value string) string { return value }
