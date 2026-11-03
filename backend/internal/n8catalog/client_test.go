package n8catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestClientListsEnabledAudioByOrganizationAndStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/center-api/material/audio/select" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["organization_id"] != "org-1" || body["status"] != "ENABLE" {
			t.Fatalf("body=%v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"id": 11, "name": "新音频", "organization_id": 1, "category_ids": []int64{3}, "status": "ENABLE", "updated_at": "2026-10-01 00:00:00"},
			{"id": 12, "name": "停用", "organization_id": 1, "category_ids": []int64{3}, "status": "DISABLE"},
		}}})
	}))
	defer server.Close()
	items, err := NewClient(server.Client(), server.URL+"/center-api", "token").ListAudio(context.Background(), "org-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 11 || items[0].OrganizationID != 1 || items[0].UpdatedAt != "2026-10-01 00:00:00" {
		t.Fatalf("items=%+v", items)
	}
}

func TestDirectoryFailureHasSafeDiagnostic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("private-token-must-not-appear"))
	}))
	defer server.Close()
	_, err := NewClient(server.Client(), server.URL, "private-token-must-not-appear").ListProducerAccounts(context.Background())
	var failure *DirectoryError
	if !errors.As(err, &failure) || failure.Kind != "http" || failure.Status != http.StatusUnauthorized {
		t.Fatalf("expected safe HTTP diagnostic, got %v", err)
	}
	if strings.Contains(err.Error(), "private-token-must-not-appear") {
		t.Fatal("diagnostic contains private data")
	}
}

func TestClientListsMaterialPageProducerAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/center-api/auth/admin_user/select" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("N8-Admin-Token") != "server-token" {
			t.Fatalf("missing server-side N8 admin token")
		}
		var body struct {
			PageSize  int `json:"page_size"`
			Filtering []struct {
				Field    string `json:"field"`
				Operator string `json:"operator"`
				Value    string `json:"value"`
			} `json:"filtering"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.PageSize != 100 || len(body.Filtering) != 0 {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"name": "无组织账号", "status": "ENABLE", "department": "DESIGN"},
			{"name": "有效账号", "status": "ENABLE", "department": "DESIGN", "organization_info": map[string]any{"0": map[string]any{"id": 1}}},
			{"name": "停用账号", "status": "DISABLE", "department": "DESIGN", "organization_info": map[string]any{"0": map[string]any{"id": 1}}},
			{"name": "其他组织账号", "status": "ENABLE", "department": "DESIGN", "organization_info": map[string]any{"0": map[string]any{"id": 2}}},
		}}})
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL+"/center-api", "server-token")
	accounts, err := client.ListProducerAccounts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(accounts, []string{"无组织账号", "有效账号", "其他组织账号"}) {
		t.Fatalf("accounts = %#v", accounts)
	}
}

func TestClientListsAllProducerAccountPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.PageSize != 100 {
			t.Fatalf("page_size = %d", body.PageSize)
		}
		items := make([]map[string]any, 0, 100)
		if body.Page == 1 {
			for i := 0; i < 100; i++ {
				items = append(items, map[string]any{"name": fmt.Sprintf("管理员%d", i), "status": "ENABLE"})
			}
		} else if body.Page == 2 {
			items = append(items, map[string]any{"name": "第二页管理员", "status": "ENABLE"})
		} else {
			t.Fatalf("unexpected page %d", body.Page)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": items}})
	}))
	defer server.Close()

	accounts, err := NewClient(server.Client(), server.URL, "server-token").ListProducerAccounts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 101 || accounts[100] != "第二页管理员" {
		t.Fatalf("accounts length/tail = %d/%q", len(accounts), accounts[len(accounts)-1])
	}
}

func TestClientSearchesWorksByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/union-api/works/select" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("N8-Admin-Token") != "server-token" {
			t.Fatalf("missing server-side N8 admin token")
		}
		var body struct {
			PageSize int    `json:"page_size"`
			Keyword  string `json:"keyword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.PageSize != 100 || body.Keyword != "小妾造反后" {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"id": 9001, "name": "小妾造反后", "cp_work_id": "cp-1", "cp_type": "小说"},
		}}})
	}))
	defer server.Close()

	client := NewWorksClient(server.Client(), server.URL+"/center-api", "server-token")
	works, err := client.SearchWorks(context.Background(), "小妾造反后")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(works, []Work{{ID: 9001, Name: "小妾造反后", CPWorkID: "cp-1", CPType: "小说"}}) {
		t.Fatalf("works = %#v", works)
	}
}

func TestClientSearchesTitleFragmentWhenFullTitleHasNoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Keyword string `json:"keyword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		list := []map[string]any{}
		if body.Keyword == "把行动都给了闺蜜" {
			list = append(list, map[string]any{"id": 989025, "name": "鼓励式老公,却把行动都给了闺蜜"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": list}})
	}))
	defer server.Close()

	client := NewWorksClient(server.Client(), server.URL, "server-token")
	works, err := client.SearchWorks(context.Background(), "激励式老公，把行动都给了闺蜜")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(works, []Work{{ID: 989025, Name: "鼓励式老公,却把行动都给了闺蜜"}}) {
		t.Fatalf("works = %#v", works)
	}
}

func TestWorksClientUsesConfiguredWorksPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"id": 7, "name": "同名作品"},
		}}})
	}))
	defer server.Close()
	client := NewWorksClientWithPath(server.Client(), server.URL, "token", "/material/work/select")
	works, err := client.SearchWorks(context.Background(), "同名作品")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/material/work/select" {
		t.Fatalf("path=%q", gotPath)
	}
	if len(works) != 1 || works[0].ID != 7 {
		t.Fatalf("works=%+v", works)
	}
}
