package giantmaterialresolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveNormalizesQingyuMaterialAndBooks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/select" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("N8-Admin-Token") != "server-token" {
			t.Fatalf("token=%q", r.Header.Get("N8-Admin-Token"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got := body["ocean_material_ids"].([]any)[0]; got != "7689285397448523826" {
			t.Fatalf("material id=%v", got)
		}
		_, _ = w.Write([]byte(`{"code":"SUCCESS","data":{"list":[{"id":"10122315","video_url":"https://material.hnqingyuwen.top/a.mp4","duration":281.03,"works":[{"cp_work_id":"748725","cp_type":"QM","name":"事不过三，过三遭殃"}]}]}}`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{Transport: rewriteTransport{target: server.URL}}, "https://n8.hnqingyuwen.top/select", "server-token")
	material, err := client.Resolve(context.Background(), "7689285397448523826")
	if err != nil {
		t.Fatal(err)
	}
	if material.MaterialID != "10122315" || material.VideoURL != "https://material.hnqingyuwen.top/a.mp4" || material.DurationSeconds != 281.03 {
		t.Fatalf("material=%+v", material)
	}
	if len(material.Books) != 1 || material.Books[0].PlatformBookID != "748725" || material.Books[0].PlatformName != "七猫" {
		t.Fatalf("books=%+v", material.Books)
	}
}

func TestResolveNormalizesChangduBookWith121PlatformID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"SUCCESS","data":{"list":[{"id":"10122316","video_url":"https://material.hnqingyuwen.top/a.mp4","works":[{"cp_work_id":"7632217270088895550","cp_type":"CD","name":"常读测试书"}]}]}}`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{Transport: rewriteTransport{target: server.URL}}, "https://n8.hnqingyuwen.top/select", "server-token")
	material, err := client.Resolve(context.Background(), "7683728935785873458")
	if err != nil {
		t.Fatal(err)
	}
	if len(material.Books) != 1 {
		t.Fatalf("books=%+v", material.Books)
	}
	book := material.Books[0]
	if book.PlatformCode != "CD" || book.PlatformName != "常读" || book.PlatformID != "2" {
		t.Fatalf("book=%+v", book)
	}
}

func TestResolveNormalizesYangguangBookToDianzhongPlatformID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"SUCCESS","data":{"list":[{"id":"10122319","video_url":"https://material.hnqingyuwen.top/a.mp4","works":[{"cp_work_id":"2053423598178718886","cp_type":"YG","name":"测试作品"}]}]}}`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{Transport: rewriteTransport{target: server.URL}}, "https://n8.hnqingyuwen.top/select", "server-token")
	material, err := client.Resolve(context.Background(), "7656746438941999123")
	if err != nil {
		t.Fatal(err)
	}
	if len(material.Books) != 1 {
		t.Fatalf("books=%+v", material.Books)
	}
	book := material.Books[0]
	if book.PlatformCode != "YG" || book.PlatformName != "点众付费" || book.PlatformID != "4" {
		t.Fatalf("book=%+v", book)
	}
}

func TestResolveNormalizesZhihuBookWith121PlatformID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"SUCCESS","data":{"list":[{"id":"10122318","video_url":"https://material.hnqingyuwen.top/a.mp4","works":[{"cp_work_id":"2053423598178718885","cp_type":"ZH","name":"为了抢风头，她重生毒死了全班"}]}]}}`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{Transport: rewriteTransport{target: server.URL}}, "https://n8.hnqingyuwen.top/select", "server-token")
	material, err := client.Resolve(context.Background(), "7656746438941999123")
	if err != nil {
		t.Fatal(err)
	}
	if len(material.Books) != 1 {
		t.Fatalf("books=%+v", material.Books)
	}
	book := material.Books[0]
	if book.PlatformCode != "ZH" || book.PlatformName != "知乎" || book.PlatformID != "15" {
		t.Fatalf("book=%+v", book)
	}
}

func TestResolveUsesTrailingBookCityTagWhenProviderCodeIsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"SUCCESS","data":{"list":[{"id":"10122317","video_url":"https://material.hnqingyuwen.top/a.mp4","works":[{"cp_work_id":"987654","cp_type":"HY","name":"公司新招的总监有纹身，我向妻子提出了离婚 (黑岩)"}]}]}}`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{Transport: rewriteTransport{target: server.URL}}, "https://n8.hnqingyuwen.top/select", "server-token")
	material, err := client.Resolve(context.Background(), "7683728935785873458")
	if err != nil {
		t.Fatal(err)
	}
	if len(material.Books) != 1 {
		t.Fatalf("books=%+v", material.Books)
	}
	book := material.Books[0]
	if book.Title != "公司新招的总监有纹身，我向妻子提出了离婚" || book.PlatformName != "黑岩" || book.PlatformID != "1" || book.PlatformCode != "HY" {
		t.Fatalf("book=%+v", book)
	}
}

type rewriteTransport struct{ target string }

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	base, _ := http.NewRequest(http.MethodGet, t.target, nil)
	clone.URL.Scheme = base.URL.Scheme
	clone.URL.Host = base.URL.Host
	return http.DefaultTransport.RoundTrip(clone)
}

func TestResolveRejectsInvalidOrUnsafeMaterial(t *testing.T) {
	client := NewClient(http.DefaultClient, "https://n8.hnqingyuwen.top/center-api/material/video/select", "token")
	if _, err := client.Resolve(context.Background(), "short"); err == nil || CodeOf(err) != "INVALID_GIANT_MATERIAL_ID" {
		t.Fatalf("invalid id error=%v", err)
	}
}
