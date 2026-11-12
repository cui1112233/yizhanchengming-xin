package httpapi

import (
 "net/http"
 "testing"
 "time"
 "qiantie/backend/internal/aicut"
)

func TestAICutProjectRatioPersistAndOwnership(t *testing.T) {
 now := time.Now().UTC()
 store := aicut.NewMemoryStore()
 store.CreateProject(t.Context(), "alice", aicut.Project{ID:"ratio-test",Title:"ratio",AspectRatio:"9:16",CreatedAt:now,UpdatedAt:now})
 mux := http.NewServeMux()
 RegisterAICutRoutes(mux,BridgeAuth{Secret:"secret",Now:func()time.Time{return now}},store,func()time.Time{return now})
 path := "/api/ai-cut/projects/ratio-test"
 good := signedJSONRequest(t,mux,now,"alice",http.MethodPatch,path,map[string]any{"aspectRatio":"16:9"})
 if good.Code != 200 {t.Fatalf("patch=%d %s",good.Code,good.Body.String())}
 loaded := signedJSONRequest(t,mux,now,"alice",http.MethodGet,path,nil)
 if !containsText(loaded.Body.String(),`"aspectRatio":"16:9"`) {t.Fatal(loaded.Body.String())}
 bad := signedJSONRequest(t,mux,now,"alice",http.MethodPatch,path,map[string]any{"aspectRatio":"bad"})
 if bad.Code != 400 {t.Fatalf("invalid=%d",bad.Code)}
 foreign := signedJSONRequest(t,mux,now,"bob",http.MethodPatch,path,map[string]any{"aspectRatio":"1:1"})
 if foreign.Code != 404 {t.Fatalf("foreign=%d",foreign.Code)}
 final,_ := store.Project(t.Context(),"alice","ratio-test")
 if final.AspectRatio != "16:9" {t.Fatal(final)}
}
