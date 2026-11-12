package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"qiantie/backend/internal/aicut"
	"strings"
)

func RegisterAICutPersonalTemplateRoutes(mux *http.ServeMux, auth BridgeAuth, store *aicut.PersonalTemplateStore) {
	if mux == nil || store == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			i, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(i.Username) == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, r, strings.TrimSpace(i.Username))
		})))
	}
	fail := func(w http.ResponseWriter, e error) {
		switch {
		case errors.Is(e, aicut.ErrInvalidPersonalTemplate):
			writeJSON(w, 400, map[string]string{"error": "模板格式无效；暂不支持复合片段、转场、联动、私有字体和人物蒙版"})
		case errors.Is(e, aicut.ErrPersonalTemplateNotFound):
			writeJSON(w, 404, map[string]string{"error": "模板不存在"})
		default:
			writeJSON(w, 500, map[string]string{"error": "模板保存或读取失败"})
		}
	}
	handle("GET /api/ai-cut/personal-templates", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := store.List(r.Context(), o)
		if e != nil {
			fail(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"templates": v})
	})
	handle("POST /api/ai-cut/personal-templates", func(w http.ResponseWriter, r *http.Request, o string) {
		var input struct {
			Title  string          `json:"title"`
			Layout json.RawMessage `json:"layout"`
		}
		if !decodeAICutJSON(w, r, &input) {
			return
		}
		l, e := aicut.DecodePersonalTemplateLayout(input.Layout)
		if e != nil {
			fail(w, e)
			return
		}
		v, e := store.Create(r.Context(), o, input.Title, l)
		if e != nil {
			fail(w, e)
			return
		}
		writeJSON(w, 201, map[string]any{"template": v})
	})
	handle("DELETE /api/ai-cut/personal-templates/{templateId}", func(w http.ResponseWriter, r *http.Request, o string) {
		if e := store.Delete(r.Context(), o, r.PathValue("templateId")); e != nil {
			fail(w, e)
			return
		}
		w.WriteHeader(204)
	})
}
