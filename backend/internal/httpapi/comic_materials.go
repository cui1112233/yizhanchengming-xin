package httpapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func serveArtifactImage(w http.ResponseWriter, r *http.Request, files *localartifact.Store, assetID string) bool {
	if files == nil || assetID == "" {
		return false
	}
	for _, extension := range []string{".png", ".jpg", ".webp"} {
		file, err := files.Open(assetID + extension)
		if err != nil {
			continue
		}
		defer file.Close()
		first := make([]byte, 512)
		n, _ := file.Read(first)
		mediaType := http.DetectContentType(first[:n])
		_, _ = file.Seek(0, io.SeekStart)
		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, r, assetID, time.Time{}, file)
		return true
	}
	return false
}

// saveComicMaterialImage validates administrator uploads with the same image
// store used by the local executor, so reference pictures are never exposed as
// arbitrary URLs or browser-only data URLs.
func saveComicMaterialImage(files *localartifact.Store, dataURL string) (string, error) {
	mediaType, payload, err := decodeImageDataURL(dataURL)
	if err != nil {
		return "", err
	}
	assetID, err := localexecutor.NewArtifactID()
	if err != nil {
		return "", err
	}
	if _, err := files.SaveImage(assetID, mediaType, bytes.NewReader(payload)); err != nil {
		return "", err
	}
	return assetID, nil
}

func writeComicMaterialUploadError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("漫画素材图片保存失败：%v", err)})
}
