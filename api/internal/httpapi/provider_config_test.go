package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/providerconfig"
)

type providerConfigFake struct { record providerconfig.Resolved; put providerconfig.PutInput }
func (f *providerConfigFake) Put(_ context.Context,input providerconfig.PutInput)(providerconfig.Record,error){f.put=input; return providerconfig.Record{ID:"cfg_1",Provider:input.Provider,MediaKind:input.MediaKind,Model:input.Model,Enabled:input.Enabled,Configured:input.APIKey!=""},nil}
func (f *providerConfigFake) Resolve(_ context.Context,owner,kind,provider string)(providerconfig.Resolved,error){if f.record.ID==""{return providerconfig.Resolved{},providerconfig.ErrNotFound}; return f.record,nil}

func TestProviderConfigPUTInjectsOwnerAndNeverReturnsAPIKey(t *testing.T){
	store:=&providerConfigFake{}
	handler:=NewProviderConfigHandler(store,func(*http.Request)(string,error){return "owner-a",nil})
	req:=httptest.NewRequest(http.MethodPut,"/api/provider-configs/video/personal_api",strings.NewReader(`{"model":"yd2.0-mini","api_key":"secret","enabled":true}`))
	rec:=httptest.NewRecorder(); handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusOK{t.Fatalf("status=%d body=%s",rec.Code,rec.Body.String())}
	if store.put.Owner!="owner-a"||store.put.MediaKind!=providerconfig.KindVideo||store.put.APIKey!="secret"{t.Fatalf("put=%#v",store.put)}
	if strings.Contains(rec.Body.String(),"secret"){t.Fatal("API key leaked in response")}
}

func TestProviderConfigGETReturnsSanitizedRecord(t *testing.T){
	store:=&providerConfigFake{record:providerconfig.Resolved{Record:providerconfig.Record{ID:"cfg_1",Provider:"personal_api",MediaKind:providerconfig.KindVideo,Model:"yd2.0-mini",Enabled:true,Configured:true},APIKey:"secret"}}
	handler:=NewProviderConfigHandler(store,func(*http.Request)(string,error){return "owner-a",nil})
	rec:=httptest.NewRecorder(); handler.ServeHTTP(rec,httptest.NewRequest(http.MethodGet,"/api/provider-configs/video/personal_api",nil))
	if rec.Code!=http.StatusOK{t.Fatalf("status=%d",rec.Code)}
	var body map[string]any
	if err:=json.Unmarshal(rec.Body.Bytes(),&body);err!=nil{t.Fatal(err)}
	if _,ok:=body["api_key"];ok{t.Fatal("api_key must never be exposed")}
	if body["model"]!="yd2.0-mini"{t.Fatalf("body=%#v",body)}
}
