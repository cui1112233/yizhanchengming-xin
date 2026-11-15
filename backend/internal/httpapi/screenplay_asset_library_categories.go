package httpapi

import (
	"net/http"
	"qiantie/backend/internal/screenplay"
)

func registerScreenplayLibraryCategories(register func(string, func(http.ResponseWriter, *http.Request, string)), service *ScreenplayAssetLibrary) {
	for _, method := range []string{"GET", "POST"} {
		register(method+" /api/comic-screenplays/asset-library/categories", func(w http.ResponseWriter, r *http.Request, actor string) {
			repo, ok := service.Repository.(screenplay.AssetLibraryCategoryRepository)
			if !ok {
				writeJSON(w, 503, map[string]string{"error": "分类存储暂不可用"})
				return
			}
			var state screenplay.AssetLibraryCategoryState
			var err error
			if r.Method == "POST" {
				var in struct {
					Action           string `json:"action"`
					Name             string `json:"name"`
					ExpectedRevision *int64 `json:"expectedRevision"`
				}
				if !decodeAssetCheckRequest(w, r, &in, false) {
					return
				}
				if in.ExpectedRevision == nil {
					writeScreenplayError(w, screenplay.ErrInvalidArgument)
					return
				}
				state, err = repo.ChangeAssetLibraryCategory(r.Context(), actor, in.Action, in.Name, *in.ExpectedRevision)
			} else {
				state, err = repo.LoadAssetLibraryCategories(r.Context(), actor)
			}
			if err != nil {
				status := screenplayErrorStatus(err)
				message := err.Error()
				if status == 500 {
					message = "分类读取或保存失败，请重试"
				}
				writeJSON(w, status, map[string]string{"error": message})
				return
			}
			writeJSON(w, 200, map[string]any{"builtin_categories": screenplay.BuiltinAssetLibraryCategories(), "custom_categories": state.CustomCategories, "revision": state.Revision})
		})
	}
}
