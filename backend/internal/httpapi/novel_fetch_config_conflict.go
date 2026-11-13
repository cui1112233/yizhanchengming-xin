package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
)

func writeWorkshopConfigSaveError(w http.ResponseWriter, err error, fallback string) {
	if errors.Is(err, novelfetchworkshop.ErrConfigConflict) {
		writeJSON(w, 409, map[string]any{"error": err.Error(), "code": "CONFIG_CONFLICT"})
		return
	}
	writeJSON(w, 500, map[string]any{"error": fallback})
}

func checkWorkshopConfigRevision(w http.ResponseWriter, settings map[string]any, expected any) bool {
	if expected == nil {
		return true
	}
	if fmt.Sprint(expected) == fmt.Sprint(settings[novelfetchworkshop.ConfigRevisionKey]) {
		return true
	}
	writeJSON(w, http.StatusConflict, map[string]any{"error": novelfetchworkshop.ErrConfigConflict.Error(), "code": "CONFIG_CONFLICT"})
	return false
}
