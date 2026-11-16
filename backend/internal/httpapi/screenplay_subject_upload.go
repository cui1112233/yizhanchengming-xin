package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
	"strconv"
	"strings"
	"time"
)

type ScreenplaySubjectUploadService struct {
	Repository screenplay.SubjectUploadRepository
	Library    *ScreenplayAssetLibrary
	Configs    *UserConfigStore
	Runner     subjectupload.Runner
	RedisDo    subjectUploadRedisDo
}

func (s *ScreenplaySubjectUploadService) configuration(ctx context.Context, owner string, resume ...bool) (subjectupload.Config, error) {
	c := subjectupload.Config{}
	if s.Configs == nil {
		return c, fmt.Errorf("请先配置真人素材上传接口")
	}
	all, e := s.Configs.GetConfig(ctx, owner)
	if e != nil {
		return c, e
	}
	m := apiIntegrationValues(apiIntegrationStoredMap(all))
	get := func(k string) string { return strings.TrimSpace(fmt.Sprint(m[k])) }
	getClean := func(k string) string {
		v := get(k)
		if v == "<nil>" {
			return ""
		}
		return v
	}
	c.Provider = getClean("realpersonChannel")
	c.ProjectName = getClean("volcanoProjectName")
	c.AccessKeyID = getClean("volcanoAccessKeyId")
	c.SecretAccessKey = getClean("volcanoSecretAccessKey")
	c.Token = getClean("seedance2Api")
	c.Region = getClean("deyunaiRegion")
	c.ImageHostURL = getClean("imageHostUrl")
	c.ImageHostToken = getClean("imageHostApiKey")
	switch c.Provider {
	case "volcengine":
		if c.ProjectName == "" {
			c.ProjectName = "default"
		}
		if c.ProjectName == "" || c.AccessKeyID == "" || c.SecretAccessKey == "" {
			return c, fmt.Errorf("请先填写火山项目名称、AccessKeyId和SecretAccessKey")
		}
	case "deyunai":
		if c.Token == "" || (c.Region != "c39" && c.Region != "c40") {
			return c, fmt.Errorf("请填写得云API密钥并选择c39或c40")
		}
	default:
		return c, fmt.Errorf("当前真人素材渠道暂未实现")
	}
	if s.Repository == nil || s.RedisDo == nil || s.Runner == nil || s.Library == nil || s.Library.Media == nil || s.Library.Media.Objects == nil {
		return c, fmt.Errorf("真人素材上传存储或租约服务未配置")
	}
	if c.ImageHostURL != "" {
		if e := mergeworker.ValidatePublicHTTPSURL(c.ImageHostURL); e != nil {
			return c, fmt.Errorf("图床必须使用公网HTTPS地址")
		}
	}
	if c.ImageHostURL == "" && s.Library.Media.ReferencesProvider == nil && (len(resume) == 0 || !resume[0]) {
		return c, fmt.Errorf("请配置TOS公网图片地址或自有图床")
	}
	return c, nil
}
func subjectUploadPublic(o screenplay.SubjectUploadOperation) map[string]any {
	v := map[string]any{"requestKey": o.RequestKey, "status": o.Status, "phase": o.Phase, "message": o.Message}
	if o.Status == "rejected" {
		v["status"] = "failed"
		v["safeToStartNew"] = true
	}
	if o.Result != nil {
		v["result"] = *o.Result
	}
	return map[string]any{"operation": v}
}

// Recovery may change a running record only after owning the same request lease.
func (s *ScreenplaySubjectUploadService) reconcileSubjectUpload(ctx context.Context, o screenplay.SubjectUploadOperation) screenplay.SubjectUploadOperation {
	if o.Status != "running" {
		return o
	}
	release, e := subjectUploadLease(ctx, s.RedisDo, o.Owner, o.RequestKey)
	if e != nil {
		return o
	}
	defer release()
	current, e := s.Repository.GetSubjectUpload(ctx, o.Owner, o.RequestKey)
	if e != nil || current.Status != "running" {
		if e == nil {
			return current
		}
		return o
	}
	current.Status = "paused"
	current.Message = "上次等待已中断，可使用原请求继续核对"
	if current.Progress.PendingAction != "" {
		current.Status = "needs_review"
		current.Message = "上次创建结果尚未确认，请人工核对，勿重复创建"
	}
	saved, e := s.Repository.SaveSubjectUpload(ctx, current, current.Revision)
	if e != nil {
		return o
	}
	return saved
}
func registerScreenplaySubjectUploads(register func(string, func(http.ResponseWriter, *http.Request, string)), s *ScreenplaySubjectUploadService) {
	register("GET /api/comic-screenplays/asset-library/subject-upload-capability", func(w http.ResponseWriter, r *http.Request, o string) {
		c, e := s.configuration(r.Context(), o)
		label := "上传真人素材"
		if c.Provider == "volcengine" {
			label = "上传火山引擎素材"
		}
		if c.Provider == "deyunai" {
			label = "上传得云AI素材"
		}
		reason := ""
		if e != nil {
			reason = e.Error()
		}
		writeJSON(w, 200, map[string]any{"provider": c.Provider, "label": label, "available": e == nil, "reason": reason})
	})
	register("GET /api/comic-screenplays/asset-library/subject-uploads/{requestKey}", func(w http.ResponseWriter, r *http.Request, o string) {
		if s.Repository == nil {
			writeJSON(w, 503, map[string]string{"error": "素材操作存储未配置"})
			return
		}
		v, e := s.Repository.GetSubjectUpload(r.Context(), o, r.PathValue("requestKey"))
		if e != nil {
			writeJSON(w, 404, map[string]string{"error": "上传操作不存在"})
			return
		}
		v = s.reconcileSubjectUpload(r.Context(), v)
		writeJSON(w, 200, subjectUploadPublic(v))
	})
	register("POST /api/comic-screenplays/asset-library/subject-uploads", func(w http.ResponseWriter, r *http.Request, o string) { s.serveUpload(w, r, o) })
}
func (s *ScreenplaySubjectUploadService) serveUpload(w http.ResponseWriter, r *http.Request, owner string) {
	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	key := ""
	preflight := true
	fail := func(code int, msg string) {
		response := map[string]any{"error": msg}
		if preflight && validAssetLibraryRequestKey(key) && s.Repository != nil && (code == 400 || code == 404 || code == 409) {
			persist, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer stop()
			rejected, err := s.Repository.SaveSubjectUpload(persist, screenplay.SubjectUploadOperation{Owner: owner, RequestKey: key, Digest: "rejected-v1", Status: "rejected", Phase: "validation", Message: msg}, 0)
			if err == nil {
				response = subjectUploadPublic(rejected)
				response["error"] = msg
			}
		}
		writeJSON(w, code, response)
	}
	c, e := s.configuration(r.Context(), owner, true)
	if e != nil {
		fail(400, e.Error())
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, screenplayLibraryImageLimit+(1<<20))
	if e = r.ParseMultipartForm(1 << 20); e != nil {
		fail(400, "请提供不超过40MiB的图片与名称")
		return
	}
	defer r.MultipartForm.RemoveAll()
	key = r.FormValue("requestKey")
	name, desc, group := strings.TrimSpace(r.FormValue("name")), r.FormValue("description"), r.FormValue("groupId")
	if !validAssetLibraryRequestKey(key) || name == "" || len(name) > 768 || len(desc) > 16384 || len(group) > 512 {
		fail(400, "请填写有效请求标识与名称")
		return
	}
	var raw []byte
	fileName := "image"
	contentType := ""
	entryID := r.FormValue("entryId")
	file, header, fileErr := r.FormFile("file")
	if fileErr == nil {
		defer file.Close()
		if entryID != "" {
			fail(400, "图片与库条目只能选择一个")
			return
		}
		raw, e = io.ReadAll(io.LimitReader(file, screenplayLibraryImageLimit+1))
		fileName = header.Filename
	} else if entryID != "" {
		entry, x := s.Library.Repository.GetAssetLibraryEntry(r.Context(), owner, entryID)
		if x != nil {
			fail(404, "库条目不存在")
			return
		}
		if entry.Type != screenplay.AssetCharacter || entry.ImageObject.ObjectKey == "" {
			fail(400, "请选择已保存图片的角色")
			return
		}
		if entry.Image.Name != "" {
			fileName = entry.Image.Name
		}
		rev, x := strconv.ParseInt(r.FormValue("expectedRevision"), 10, 64)
		if x != nil || rev != entry.Revision {
			fail(409, "库条目已变化")
			return
		}
		obj := agentworkspace.Asset{ObjectKey: entry.ImageObject.ObjectKey, ContentType: entry.ImageObject.ContentType, ByteSize: entry.ImageObject.ByteSize}
		reader, x := s.Library.Media.Objects.Open(r.Context(), obj)
		if x != nil {
			fail(400, "读取库图片失败")
			return
		}
		raw, e = io.ReadAll(io.LimitReader(reader, screenplayLibraryImageLimit+1))
		reader.Close()
	} else {
		fail(400, "请选择图片")
		return
	}
	if e != nil || len(raw) == 0 || len(raw) > screenplayLibraryImageLimit {
		fail(400, "图片不能为空且不能超过40MiB")
		return
	}
	contentType = http.DetectContentType(raw)
	ext := map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/gif": "gif", "image/webp": "webp"}[contentType]
	if ext == "" {
		fail(400, "请选择PNG、JPEG、GIF或WebP图片")
		return
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(raw)); err != nil {
		fail(400, "图片内容无效，请重新选择图片")
		return
	}
	credentialFingerprint := sha256.Sum256([]byte(c.AccessKeyID + "\x00" + c.SecretAccessKey + "\x00" + c.Token))
	payload, _ := json.Marshal([]string{c.Provider, c.ProjectName, c.Region, name, desc, group, fileName, fmt.Sprintf("%x", credentialFingerprint), fmt.Sprintf("%x", sha256.Sum256(raw))})
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	preflight = false
	release, e := subjectUploadLease(r.Context(), s.RedisDo, owner, key)
	if e != nil {
		fail(409, "上传操作正在执行或租约服务不可用，请核对原请求")
		return
	}
	defer release()
	op, e := s.Repository.GetSubjectUpload(r.Context(), owner, key)
	if e == nil {
		if op.Status == "rejected" {
			writeJSON(w, 200, subjectUploadPublic(op))
			return
		}
		if op.Digest != digest {
			fail(409, "同一请求标识不能更换内容")
			return
		}
		if op.Status == "succeeded" || op.Progress.PendingAction != "" || op.Status == "needs_review" {
			if op.Progress.PendingAction != "" && op.Status != "needs_review" {
				op.Status = "needs_review"
				op.Message = "创建结果尚未确认，请人工核对"
				op, _ = s.Repository.SaveSubjectUpload(r.Context(), op, op.Revision)
			}
			writeJSON(w, 200, subjectUploadPublic(op))
			return
		}
	} else if !errors.Is(e, screenplay.ErrNotFound) {
		fail(503, "读取上传记录失败")
		return
	} else {
		imageID := bulkLibraryID(owner, key, digest, 0, "subject-source")
		objectKey, x := agentworkspace.AgentAssetObjectKey(owner, screenplayLibraryObjectProject, imageID, ext)
		if x != nil {
			fail(400, "无法保存图片")
			return
		}
		source := agentworkspace.Asset{ID: imageID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: "subject." + ext, ContentType: contentType, ObjectKey: objectKey, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
		if x = s.Library.Media.Objects.Put(r.Context(), source, bytes.NewReader(raw)); x != nil {
			fail(503, "保存源图片失败")
			return
		}
		op = screenplay.SubjectUploadOperation{Owner: owner, RequestKey: key, Digest: digest, Status: "paused", Phase: "prepared", Source: source}
		op, e = s.Repository.SaveSubjectUpload(r.Context(), op, 0)
		if e != nil {
			fail(409, "保存上传记录失败，请核对原请求")
			return
		}
	}
	save := func() error {
		persist, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer stop()
		v, x := s.Repository.SaveSubjectUpload(persist, op, op.Revision)
		if x == nil {
			op = v
		}
		return x
	}
	input := subjectupload.Input{Name: name, Description: desc, GroupID: group, Image: raw, ContentType: contentType, FileName: fileName}
	input.ImageURL = op.Progress.ImageURL
	if op.Progress.AssetID == "" && input.ImageURL == "" && s.Library.Media.ReferencesProvider != nil {
		u, x := s.Library.Media.ReferencesProvider.ReferenceURL(ctx, op.Source, 10*time.Minute)
		if x == nil && mergeworker.ValidatePublicHTTPSURL(u) == nil {
			input.ImageURL = u
		}
	}
	if op.Progress.AssetID == "" && input.ImageURL == "" && c.ImageHostURL == "" {
		op.Status = "paused"
		op.Message = "图片准备已暂停，请配置TOS公网地址或自有图床后使用原请求继续"
		_ = save()
		writeJSON(w, 200, subjectUploadPublic(op))
		return
	}
	op.Status = "running"
	op.Message = "正在上传真人素材"
	if e = save(); e != nil {
		fail(409, "保存进度失败")
		return
	}
	result, runErr := s.Runner.Run(ctx, c, input, op.Progress, func(_ context.Context, p subjectupload.Progress) error {
		op.Progress = p
		op.Phase = p.PendingAction
		if op.Phase == "" {
			op.Phase = p.Status
		}
		return save()
	})
	if runErr == nil {
		op.Status = "succeeded"
		op.Phase = "completed"
		op.Message = "真人素材上传成功"
		op.Result = &result
	} else {
		op.Status = "failed"
		op.Message = "上传未完成，请核对原请求"
		var providerError *subjectupload.RunError
		if errors.As(runErr, &providerError) {
			op.Progress = providerError.Progress
			if providerError.Kind == "paused" || providerError.Kind == "needs_review" {
				op.Status = providerError.Kind
			}
			op.Message = providerError.Message
		}
		if ctx.Err() != nil {
			op.Status = "paused"
			op.Message = "等待已停止，远端创建结果请核对原请求"
		}
		if op.Progress.PendingAction != "" {
			op.Status = "needs_review"
			op.Message = "创建结果尚未确认，请人工核对，勿重复创建"
		}
	}
	if e = save(); e != nil {
		fail(503, "保存结果失败，请核对原请求")
		return
	}
	writeJSON(w, 200, subjectUploadPublic(op))
}
