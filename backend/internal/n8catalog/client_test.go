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

func TestClientListsAudioWhenN8ReturnsScalarCategoryID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"id": 33, "name": "泡沫.MP3", "organization_id": 1, "category_ids": 8, "status": "ENABLE"},
		}}})
	}))
	defer server.Close()

	items, err := NewClient(server.Client(), server.URL+"/center-api", "token").ListAudio(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || !reflect.DeepEqual(items[0].CategoryIDs, []int64{8}) {
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
		if body.PageSize != 100 || len(body.Filtering) != 1 || body.Filtering[0].Field != "name" || body.Filtering[0].Operator != "LIKE" || body.Filtering[0].Value != "小妾造反后" {
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
			Filtering []struct {
				Value string `json:"value"`
			} `json:"filtering"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		list := []map[string]any{}
		if len(body.Filtering) == 1 && body.Filtering[0].Value == "把行动都给了闺蜜" {
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

func TestWorksClientSearchesByCPWorkID(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&payload)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"id": 1046266, "name": "她为发小有求必应，我转身离开", "cp_work_id": "793433", "cp_type": "QM"},
		}}})
	}))
	defer server.Close()
	client := NewWorksClientWithPath(server.Client(), server.URL, "token", "/works/select")
	works, err := client.SearchWorksByCPWorkID(context.Background(), "793433")
	if err != nil {
		t.Fatal(err)
	}
	filters, _ := payload["filtering"].([]any)
	if len(filters) != 1 {
		t.Fatalf("payload=%v", payload)
	}
	filter, _ := filters[0].(map[string]any)
	if filter["field"] != "cp_work_id" || filter["value"] != "793433" {
		t.Fatalf("filter=%v", filter)
	}
	if len(works) != 1 || works[0].ID != 1046266 || works[0].CPWorkID != "793433" {
		t.Fatalf("works=%+v", works)
	}
}

func TestClientListsOrganizationsFromNestedTree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/center-api/auth/organization/tree" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("N8-Admin-Token") != "server-token" {
			t.Fatalf("missing server-side N8 admin token")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "message": "成功", "data": []map[string]any{
			{
				"id": 1, "name": "总部", "status": "ENABLE",
				"children": []map[string]any{
					{"id": 11, "name": "制作一组", "status": "ENABLE", "children": []any{}},
					{"id": 12, "name": "制作二组", "status": "ENABLE", "children": []any{}},
				},
			},
		}})
	}))
	defer server.Close()

	organizations, err := NewClient(server.Client(), server.URL+"/center-api", "server-token").ListOrganizations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(organizations, []Organization{{ID: "1", Name: "总部"}, {ID: "11", Name: "制作一组"}, {ID: "12", Name: "制作二组"}}) {
		t.Fatalf("organizations = %#v", organizations)
	}
}

func TestClientListsEnabledAudioCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/center-api/material/category/select" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "message": "成功", "data": map[string]any{"list": []map[string]any{
			{"id": 12, "name": "都市", "status": "ENABLE"},
			{"id": 13, "name": "古风", "status": "DISABLE"},
		}}})
	}))
	defer server.Close()

	categories, err := NewClient(server.Client(), server.URL+"/center-api", "server-token").ListAudioCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 || categories[0].ID != 12 || categories[0].Name != "都市" {
		t.Fatalf("categories = %#v", categories)
	}
}

func TestClientExposesUpstreamRejectionCodeWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "UN_LOGIN", "message": "登录已失效", "data": nil})
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL+"/center-api/auth", "server-token")
	_, err := client.ListOrganizations(context.Background())
	var failure *DirectoryError
	if !errors.As(err, &failure) {
		t.Fatalf("expected DirectoryError, got %v", err)
	}
	if failure.Kind != "service_rejected" || failure.Code != "UN_LOGIN" || failure.Message != "登录已失效" {
		t.Fatalf("failure = %+v", failure)
	}
	if strings.Contains(err.Error(), "server-token") {
		t.Fatal("diagnostic contains private data")
	}
}

func TestClientReadsCamelCaseOrganizationInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "SUCCESS", "data": map[string]any{"list": []map[string]any{
			{"name": "驼峰组织账号", "status": "ENABLE", "organizationInfo": []map[string]any{{"id": 1}, {"id": 11}}},
			{"name": "下划线组织账号", "status": "ENABLE", "organization_info": []map[string]any{{"id": 2}}},
		}}})
	}))
	defer server.Close()

	entries, err := NewClient(server.Client(), server.URL+"/center-api", "server-token").ListProducerAccountDirectory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %#v", entries)
	}
	if !reflect.DeepEqual(entries[0].OrganizationIDs, []string{"1", "11"}) || !reflect.DeepEqual(entries[1].OrganizationIDs, []string{"2"}) {
		t.Fatalf("organization bindings = %#v", entries)
	}
}
