package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

var novelFetchBookIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var novelFetchVersionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

type NovelFetchBridgeAuth struct {
	Secret    string
	Now       func() time.Time
	MaxSkew   time.Duration
	Users     BridgeUserResolver
	AuthStore auth.Store
}

func novelFetchCanonicalPayload(username, issuedAt, owner, method, pathname string) string {
	return strings.Join([]string{username, issuedAt, owner, method, pathname}, "\n")
}

func novelFetchSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (a NovelFetchBridgeAuth) authenticate(req *http.Request) (BridgeIdentity, error) {
	if token := bearerToken(req); token != "" && a.AuthStore != nil {
		now := time.Now
		if a.Now != nil {
			now = a.Now
		}
		session, err := a.AuthStore.FindSession(req.Context(), token, now().UTC())
		if err != nil {
			return BridgeIdentity{}, errors.New("invalid login session")
		}
		account, err := a.AuthStore.FindAccount(req.Context(), session.Username)
		if err != nil || !account.Active {
			return BridgeIdentity{}, errors.New("inactive login account")
		}
		return BridgeIdentity{Username: account.Username, IsOwner: account.IsOwner}, nil
	}
	if strings.TrimSpace(a.Secret) == "" {
		return BridgeIdentity{}, errors.New("bridge secret is not configured")
	}
	username := strings.TrimSpace(req.Header.Get(HeaderUsername))
	ownerRaw := req.Header.Get(HeaderIsOwner)
	issuedRaw := req.Header.Get(HeaderIssuedAt)
	signature := strings.TrimSpace(req.Header.Get(HeaderSignature))
	if username == "" || ownerRaw == "" || issuedRaw == "" || signature == "" {
		return BridgeIdentity{}, errors.New("missing signed bridge headers")
	}
	isOwner, err := strconv.ParseBool(ownerRaw)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid owner header")
	}
	issuedUnix, err := strconv.ParseInt(issuedRaw, 10, 64)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid issued-at header")
	}
	now := time.Now
	if a.Now != nil {
		now = a.Now
	}
	maxSkew := 5 * time.Minute
	if a.MaxSkew > 0 {
		maxSkew = a.MaxSkew
	}
	delta := now().Sub(time.Unix(issuedUnix, 0))
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return BridgeIdentity{}, errors.New("expired bridge signature")
	}
	expected := novelFetchSignature(a.Secret, novelFetchCanonicalPayload(username, issuedRaw, ownerRaw, req.Method, req.URL.Path))
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return BridgeIdentity{}, errors.New("malformed bridge signature")
	}
	expectedBytes, _ := hex.DecodeString(expected)
	if !hmac.Equal(provided, expectedBytes) {
		return BridgeIdentity{}, errors.New("invalid bridge signature")
	}
	if a.Users != nil {
		if err := a.Users.ResolveBridgeUser(req.Context(), username, isOwner); err != nil {
			return BridgeIdentity{}, err
		}
	}
	return BridgeIdentity{Username: username, IsOwner: isOwner}, nil
}

func (a NovelFetchBridgeAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, err := a.authenticate(req)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, identity)))
	})
}

func novelFetchBodyPathValues(w http.ResponseWriter, req *http.Request) (string, string, bool) {
	bookID := strings.TrimSpace(req.PathValue("bookId"))
	if !novelFetchBookIDPattern.MatchString(bookID) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
		return "", "", false
	}
	versionID := strings.TrimSpace(req.PathValue("versionId"))
	if versionID != "" && !novelFetchVersionIDPattern.MatchString(versionID) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的版本ID"})
		return "", "", false
	}
	return bookID, versionID, true
}

func registerNovelFetchWorkshopRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, aiJudge novelfetchworkshop.AIGenderJudge) {
	mux.HandleFunc("GET /api/novel-fetch-workshop/runs", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.URL.Query().Get("bookId"))
		if bookID != "" && !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		runs, err := store.ListRuns(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "获取执行审计失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/runs", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var record novelfetchworkshop.RunRecord
		if err := decodeNovelFetchJSON(w, req, &record); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		if !novelFetchBookIDPattern.MatchString(strings.TrimSpace(record.BookID)) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		if err := store.PutRun(req.Context(), identity.Username, record); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "执行审计字段无效"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		tasks, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "获取任务列表失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}/bodies", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, _, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		bodies, err := store.ListBodyRefs(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文列表失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"bodies": bodies})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		body, err := store.GetBody(req.Context(), identity.Username, bookID, versionID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "正文不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, body)
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		var input struct {
			Content string `json:"content"`
			State   string `json:"state"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		ref, err := store.PutBody(req.Context(), identity.Username, novelfetchworkshop.BodyRecord{
			BookID:  bookID,
			BodyRef: novelfetchworkshop.BodyRef{VersionID: versionID, State: input.State},
			Content: input.Content,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, ref)
	})

	mux.HandleFunc("DELETE /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		deleted, err := store.DeleteBody(req.Context(), identity.Username, bookID, versionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		document, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "任务不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, document)
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/tasks/{bookId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		var document novelfetchworkshop.Document
		if err := decodeNovelFetchJSON(w, req, &document); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		explicitManualGender := strings.TrimSpace(document.Gender) != ""
		document.BookID = bookID
		existing, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if err != nil && !errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		document = mergeNovelFetchDocument(existing, document)
		if explicitManualGender {
			decision := novelfetchworkshop.ResolveGender(req.Context(), document.Gender, document.BookInfo, aiJudge)
			document.Gender = decision.Gender
			document.GenderSource = decision.Source
			document.GenderReason = decision.Reason
		}
		if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("DELETE /api/novel-fetch-workshop/tasks", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var body struct {
			IDs []string `json:"ids"`
		}
		if err := decodeNovelFetchJSON(w, req, &body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		ids := make([]string, 0, len(body.IDs))
		for _, raw := range body.IDs {
			id := strings.TrimSpace(raw)
			if !novelFetchBookIDPattern.MatchString(id) {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
				return
			}
			ids = append(ids, id)
		}
		result, err := store.DeleteDocuments(req.Context(), identity.Username, ids)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/platforms", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取平台表失败"})
			return
		}
		platforms := make([]map[string]any, 0)
		if configured, ok := settings["platforms"].([]any); ok {
			for _, raw := range configured {
				platform, ok := raw.(map[string]any)
				if !ok || platform["visible"] == false {
					continue
				}
				id, idOK := platform["id"].(string)
				name, nameOK := platform["name"].(string)
				if idOK && strings.TrimSpace(id) != "" && nameOK && strings.TrimSpace(name) != "" {
					platforms = append(platforms, map[string]any{"id": id, "name": name})
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"platforms": platforms})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/config", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
			return
		}
		appConfig := settings
		if configured, ok := settings["appConfig"].(map[string]any); ok {
			appConfig = configured
		}
		platforms, ok := settings["platforms"].([]any)
		if !ok {
			platforms = []any{}
		}
		styles, ok := settings["styles"].([]any)
		if !ok {
			styles = []any{}
		}
		aiConfig, _ := settings["aiConfig"].(map[string]any)
		if aiConfig == nil {
			if ai, ok := appConfig["ai"].(map[string]any); ok {
				aiConfig = map[string]any{"ai": ai}
			} else {
				aiConfig = map[string]any{}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"appConfig": appConfig,
			"platforms": platforms,
			"styles":    styles,
			"aiConfig":  aiConfig,
			"settings":  settings,
		})
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/config", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var patch map[string]any
		if err := decodeNovelFetchJSON(w, req, &patch); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		settings, _ := patch["settings"].(map[string]any)
		if settings == nil {
			settings = patch
		}
		current, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
			return
		}
		merged := mergeWorkshopConfig(current, settings)
		if err := store.PutConfig(req.Context(), identity.Username, merged); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "settings": merged})
	})
}

func mergeWorkshopConfig(current, patch map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range patch {
		currentMap, currentOK := merged[key].(map[string]any)
		patchMap, patchOK := value.(map[string]any)
		if currentOK && patchOK {
			merged[key] = mergeWorkshopConfig(currentMap, patchMap)
			continue
		}
		merged[key] = value
	}
	return merged
}

func mergeNovelFetchDocument(existing, incoming novelfetchworkshop.Document) novelfetchworkshop.Document {
	if strings.TrimSpace(incoming.Gender) == "" {
		incoming.Gender = existing.Gender
	}
	if incoming.GenderSource == "" {
		incoming.GenderSource = existing.GenderSource
	}
	if strings.TrimSpace(incoming.GenderReason) == "" {
		incoming.GenderReason = existing.GenderReason
	}
	if strings.TrimSpace(incoming.BookInfo.Category) == "" {
		incoming.BookInfo.Category = existing.BookInfo.Category
	}
	if strings.TrimSpace(incoming.BookInfo.Genre) == "" {
		incoming.BookInfo.Genre = existing.BookInfo.Genre
	}
	if incoming.Meta == nil {
		incoming.Meta = existing.Meta
	}
	if incoming.BodyRefs == nil {
		incoming.BodyRefs = existing.BodyRefs
	}
	if strings.TrimSpace(incoming.Original) == "" {
		incoming.Original = existing.Original
	}
	if strings.TrimSpace(incoming.OriginalRaw) == "" {
		incoming.OriginalRaw = existing.OriginalRaw
	}
	if strings.TrimSpace(incoming.ProcessedText) == "" {
		incoming.ProcessedText = existing.ProcessedText
	}
	if incoming.OriginalRawChars == 0 {
		incoming.OriginalRawChars = existing.OriginalRawChars
	}
	if incoming.ProcessedChars == 0 {
		incoming.ProcessedChars = existing.ProcessedChars
	}
	if incoming.MaxTxt == 0 {
		incoming.MaxTxt = existing.MaxTxt
	}
	if incoming.Versions == nil {
		incoming.Versions = existing.Versions
	}
	if incoming.Logs == nil {
		incoming.Logs = existing.Logs
	}
	return incoming
}

func decodeNovelFetchJSON(w http.ResponseWriter, req *http.Request, target any) error {
	req.Body = http.MaxBytesReader(w, req.Body, 64<<20)
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(target)
}
