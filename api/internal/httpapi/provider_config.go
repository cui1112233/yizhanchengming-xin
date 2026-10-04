package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/providerconfig"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/videogen"
)

type ProviderConfigAPI interface {
	Put(context.Context, providerconfig.PutInput) (providerconfig.Record, error)
	Resolve(context.Context, string, string, string) (providerconfig.Resolved, error)
}

type providerConfigHandler struct {
	store ProviderConfigAPI
	owner OwnerResolver
}

func NewProviderConfigHandler(store ProviderConfigAPI, owner OwnerResolver) http.Handler {
	return &providerConfigHandler{store: store, owner: owner}
}

type providerConfigPutRequest struct {
	Model     string          `json:"model"`
	APIKey    string          `json:"api_key"`
	CreateURL string          `json:"create_url,omitempty"`
	TasksURL  string          `json:"tasks_url,omitempty"`
	ResultURL string          `json:"result_url,omitempty"`
	Settings  json.RawMessage `json:"settings,omitempty"`
	Enabled   *bool           `json:"enabled,omitempty"`
}

func (h *providerConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.store == nil || h.owner == nil { http.Error(w,"provider config unavailable",http.StatusServiceUnavailable); return }
	owner, err := h.owner(r)
	if err != nil || strings.TrimSpace(owner)=="" { http.Error(w,"unauthorized",http.StatusUnauthorized); return }
	owner = strings.TrimSpace(owner)
	prefix := "/api/provider-configs/video/"
	if !strings.HasPrefix(r.URL.Path,prefix) { http.NotFound(w,r); return }
	provider := videogen.NormalizeProvider(strings.Trim(strings.TrimPrefix(r.URL.Path,prefix),"/"))
	if provider=="" { http.NotFound(w,r); return }

	switch r.Method {
	case http.MethodGet:
		resolved, err := h.store.Resolve(r.Context(),owner,providerconfig.KindVideo,provider)
		if errors.Is(err,providerconfig.ErrNotFound) { http.NotFound(w,r); return }
		if err != nil { http.Error(w,"load provider config",http.StatusInternalServerError); return }
		providerWriteJSON(w,http.StatusOK,resolved.Record)
	case http.MethodPut:
		var body providerConfigPutRequest
		decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,64<<10))
		decoder.DisallowUnknownFields()
		if err:=decoder.Decode(&body);err!=nil { http.Error(w,"invalid json",http.StatusBadRequest); return }
		var extra any
		if err:=decoder.Decode(&extra);!errors.Is(err,io.EOF) { http.Error(w,"invalid trailing json",http.StatusBadRequest); return }
		enabled:=true
		if body.Enabled!=nil { enabled=*body.Enabled }
		remote := provider != videogen.ProviderDoubaoLocal
		configured := strings.TrimSpace(body.APIKey)!=""
		if remote && !configured {
			existing, resolveErr := h.store.Resolve(r.Context(),owner,providerconfig.KindVideo,provider)
			if resolveErr != nil || !existing.Configured { http.Error(w,"api_key is required for first configuration",http.StatusBadRequest); return }
			configured = true
		}
		result,err:=h.store.Put(r.Context(),providerconfig.PutInput{Owner:owner,MediaKind:providerconfig.KindVideo,Provider:provider,Model:body.Model,APIKey:body.APIKey,CreateURL:body.CreateURL,TasksURL:body.TasksURL,ResultURL:body.ResultURL,Settings:body.Settings,Enabled:enabled})
		if err!=nil { http.Error(w,"invalid provider configuration",http.StatusBadRequest); return }
		result.Configured = configured
		providerWriteJSON(w,http.StatusOK,result)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func providerWriteJSON(w http.ResponseWriter,status int,value any){
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
