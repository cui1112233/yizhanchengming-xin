package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/novelcreation"
)

func registerNovelStudioRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), decode func(http.ResponseWriter, *http.Request, any) bool, service *novelcreation.Service, starter NovelCreationRunStarter) {
	handle("GET /api/novel-creation/studio-sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessions, err := service.ListStudioSessions(req.Context(), owner)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	})
	handle("POST /api/novel-creation/studio-sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input novelcreation.StudioCreateSessionInput
		if !decode(w, req, &input) {
			return
		}
		session, err := service.CreateStudioSession(req.Context(), owner, input)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"session": session})
	})
	handle("GET /api/novel-creation/studio-sessions/{studioID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		snapshot, err := service.GetStudioSnapshot(req.Context(), owner, req.PathValue("studioID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
	})
	handle("POST /api/novel-creation/studio-sessions/{studioID}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input novelcreation.StudioSendInput
		if !decode(w, req, &input) {
			return
		}
		snapshot, err := service.SendStudioMessage(req.Context(), owner, req.PathValue("studioID"), input)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
	})
	handle("POST /api/novel-creation/studio-sessions/{studioID}/actions/confirm", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input novelcreation.StudioActionInput
		if !decode(w, req, &input) {
			return
		}
		result, err := service.ConfirmStudioAction(req.Context(), owner, req.PathValue("studioID"), input)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, result.Run.ID)
		}
		writeJSON(w, http.StatusOK, result)
	})
	handle("POST /api/novel-creation/studio-sessions/{studioID}/actions/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ActionID string `json:"actionId"`
		}
		if !decode(w, req, &input) {
			return
		}
		result, err := service.CancelStudioAction(req.Context(), owner, req.PathValue("studioID"), novelcreation.StudioActionInput{ActionID: input.ActionID})
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
}
