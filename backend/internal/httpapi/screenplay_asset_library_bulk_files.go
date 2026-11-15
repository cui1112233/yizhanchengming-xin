package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"qiantie/backend/internal/characterarchive"
	"qiantie/backend/internal/screenplay"
)

const libraryBulkTotalLimit = 256 << 20
const libraryBulkCharacterLimit = 10000

type libraryBulkFiles struct {
	RequestKey string
	Files      map[string][]byte
	Archive    string
}

func libraryBulkError(message string) error {
	return fmt.Errorf("%w: %s", screenplay.ErrInvalidArgument, message)
}
func libraryRelativeName(name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") || path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", libraryBulkError("档案资源路径不安全")
	}
	return name, nil
}
func parseLibraryBulkFiles(w http.ResponseWriter, r *http.Request) (libraryBulkFiles, error) {
	in := libraryBulkFiles{Files: map[string][]byte{}}
	r.Body = http.MaxBytesReader(w, r.Body, libraryBulkTotalLimit+(1<<20))
	reader, err := r.MultipartReader()
	if err != nil {
		return in, libraryBulkError("请上传档案文件夹")
	}
	total := 0
	count := 0
	hasKey := false
	for {
		part, e := reader.NextPart()
		if e == io.EOF {
			break
		}
		if e != nil {
			return in, libraryBulkError("档案超过256MB或上传未完成")
		}
		if part.FormName() == "requestKey" {
			if hasKey {
				part.Close()
				return in, screenplay.ErrInvalidArgument
			}
			hasKey = true
			raw, e := io.ReadAll(io.LimitReader(part, 192))
			part.Close()
			if e != nil || !validAssetLibraryRequestKey(string(raw)) {
				return in, screenplay.ErrInvalidArgument
			}
			in.RequestKey = string(raw)
			continue
		}
		if part.FormName() != "files" {
			part.Close()
			return in, screenplay.ErrInvalidArgument
		}
		_, params, e := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if e != nil {
			part.Close()
			return in, e
		}
		name, e := libraryRelativeName(params["filename"])
		if e != nil {
			part.Close()
			return in, e
		}
		if _, exists := in.Files[name]; exists {
			part.Close()
			return in, libraryBulkError("资源文件名重复")
		}
		count++
		if count > 10000 {
			part.Close()
			return in, libraryBulkError("一次最多上传10000个文件")
		}
		raw, e := io.ReadAll(io.LimitReader(part, characterarchive.DefaultMaxInputBytes+1))
		part.Close()
		total += len(raw)
		if e != nil || int64(len(raw)) > characterarchive.DefaultMaxInputBytes || total > libraryBulkTotalLimit {
			return in, libraryBulkError("单档案最大128MB，整批最大256MB")
		}
		in.Files[name] = raw
	}
	if !hasKey {
		return in, screenplay.ErrInvalidArgument
	}
	encrypted := []string{}
	plain := []string{}
	for name := range in.Files {
		switch path.Base(name) {
		case "characters.lschar":
			encrypted = append(encrypted, name)
		case "characters.json":
			plain = append(plain, name)
		}
	}
	candidates := plain
	if len(encrypted) > 0 {
		candidates = encrypted
	}
	if len(candidates) != 1 {
		return in, libraryBulkError("请选择包含唯一characters.lschar或characters.json的文件夹")
	}
	in.Archive = candidates[0]
	return in, nil
}
func (in libraryBulkFiles) records() ([]map[string]json.RawMessage, error) {
	raw := in.Files[in.Archive]
	var err error
	if path.Ext(in.Archive) == ".lschar" {
		raw, err = characterarchive.DecodeJSON(raw)
		if err != nil {
			return nil, libraryBulkError("加密档案损坏或超过限制，请重新选择")
		}
	}
	if !utf8.Valid(raw) {
		return nil, libraryBulkError("档案文字编码不正确")
	}
	var entries []map[string]json.RawMessage
	if json.Unmarshal(raw, &entries) != nil || len(entries) == 0 {
		return nil, libraryBulkError("档案必须是非空角色数组")
	}
	if len(entries) > libraryBulkCharacterLimit {
		return nil, libraryBulkError(fmt.Sprintf("一次最多导入%d个角色", libraryBulkCharacterLimit))
	}
	for _, e := range entries {
		if e == nil {
			return nil, libraryBulkError("档案包含非法角色")
		}
	}
	return entries, nil
}
func libraryArchiveString(entry map[string]json.RawMessage, key string) (string, error) {
	v, ok := entry[key]
	if !ok || string(v) == "null" {
		return "", nil
	}
	var text string
	if json.Unmarshal(v, &text) != nil {
		return "", libraryBulkError("角色字段" + key + "必须为文字")
	}
	return text, nil
}
func (in libraryBulkFiles) resource(value string, limit int) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	if strings.HasPrefix(value, "data:") {
		head, body, ok := strings.Cut(value, ",")
		if !ok || !strings.HasSuffix(head, ";base64") {
			return nil, libraryBulkError("资源需使用base64内联格式")
		}
		if base64.StdEncoding.DecodedLen(len(body)) > limit+2 {
			return nil, libraryBulkError("资源超过大小限制")
		}
		raw, err := base64.StdEncoding.Strict().DecodeString(body)
		if err != nil || len(raw) == 0 || len(raw) > limit {
			return nil, libraryBulkError("内联资源无效或超过大小限制")
		}
		return raw, nil
	}
	name, err := libraryRelativeName(value)
	if err != nil {
		return nil, err
	}
	resolved := path.Join(path.Dir(in.Archive), name)
	raw, ok := in.Files[resolved]
	if !ok {
		return nil, libraryBulkError("缺少档案引用的资源文件：" + name)
	}
	if len(raw) == 0 || len(raw) > limit {
		return nil, libraryBulkError("资源为空或超过大小限制")
	}
	return raw, nil
}
func libraryBulkImageFormat(raw []byte) (string, string, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 {
		return "", "", libraryBulkError("档案图片无效")
	}
	ext := map[string]string{"png": "png", "jpeg": "jpg", "webp": "webp", "gif": "gif"}[format]
	ct := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp", "gif": "image/gif"}[format]
	if ext == "" {
		return "", "", libraryBulkError("档案图片格式不支持")
	}
	return ct, ext, nil
}
func libraryBulkSortedNames(in libraryBulkFiles) []string {
	names := make([]string, 0, len(in.Files))
	for name := range in.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
