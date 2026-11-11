package httpapi

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type pluginToolRunInput struct {
	Input    string `json:"input"`
	ModelID  string `json:"modelId"`
	PromptID string `json:"promptId"`
	Raw      bool   `json:"raw"`
}

type coverPromptInput struct {
	Content        string `json:"content"`
	DramaName      string `json:"dramaName"`
	PromptID       string `json:"promptId"`
	AspectRatio    string `json:"aspectRatio"`
	FontStyle      string `json:"fontStyle"`
	ReferenceCount int    `json:"referenceCount"`
}

type coverDesignInput struct {
	Content        string `json:"content"`
	DramaName      string `json:"dramaName"`
	PromptID       string `json:"promptId"`
	AspectRatio    string `json:"aspectRatio"`
	FontStyle      string `json:"fontStyle"`
	ReferenceCount int    `json:"referenceCount"`
	ModelID        string `json:"modelId"`
}

func registerPluginToolRoutes(root *http.ServeMux, auth BridgeAuth, provider batchfactoryv11.DirectorProvider, executors ...AgentTextCompletionExecutor) {
	var execute AgentTextCompletionExecutor
	if len(executors) > 0 {
		execute = executors[0]
	}
	registerAISegmentRoutes(root, auth, provider, execute)
	root.Handle("GET /api/plugins/{toolID}/config", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		items, ok := prompt.PluginToolPromptOptions(r.PathValue("toolID"))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "插件工具不存在"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"prompts": items})
	})))
	root.Handle("POST /api/plugins/cover-generation/prompt", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input coverPromptInput
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		input.Content = strings.TrimSpace(input.Content)
		input.DramaName = strings.TrimSpace(input.DramaName)
		input.PromptID = strings.TrimSpace(input.PromptID)
		input.AspectRatio = strings.TrimSpace(input.AspectRatio)
		input.FontStyle = strings.TrimSpace(input.FontStyle)
		if input.Content == "" && input.DramaName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入短剧名或文案/简介"})
			return
		}
		if input.ReferenceCount < 0 {
			input.ReferenceCount = 0
		} else if input.ReferenceCount > 20 {
			input.ReferenceCount = 20
		}
		built, ok := prompt.BuildCoverPrompt(input.PromptID, input.Content, input.DramaName, input.AspectRatio, input.FontStyle, input.ReferenceCount)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择有效的提示词"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"prompt": built})
	})))
	// /design 用 V88 自己的文本模型执行「一键封面指令」，产出最终图片提示词；
	// /prompt 只负责指令组装与预览，不调用模型。
	root.Handle("POST /api/plugins/cover-generation/design", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input coverDesignInput
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		input.Content = strings.TrimSpace(input.Content)
		input.DramaName = strings.TrimSpace(input.DramaName)
		input.PromptID = strings.TrimSpace(input.PromptID)
		input.AspectRatio = strings.TrimSpace(input.AspectRatio)
		input.FontStyle = strings.TrimSpace(input.FontStyle)
		input.ModelID = strings.TrimSpace(input.ModelID)
		if input.Content == "" && input.DramaName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入短剧名或文案/简介"})
			return
		}
		if input.ReferenceCount < 0 {
			input.ReferenceCount = 0
		} else if input.ReferenceCount > 20 {
			input.ReferenceCount = 20
		}
		if input.PromptID == "" {
			input.PromptID = "cover-design"
		}
		system, ok := prompt.PluginToolSystemPrompt("cover-generation", input.PromptID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择有效的提示词"})
			return
		}
		request := batchfactoryv11.TextCompletionRequest{
			SystemPrompt:        system,
			UserPrompt:          prompt.BuildCoverRequirement(input.Content, input.DramaName, input.FontStyle, input.AspectRatio, input.ReferenceCount),
			MaxTokens:           12000,
			Temperature:         0.4,
			DisableJSONResponse: true,
		}
		var result string
		var err error
		if execute != nil && input.ModelID != "" {
			identity, _ := BridgeIdentityFromContext(r.Context())
			result, err = execute(r.Context(), identity.Username, input.ModelID, request)
		} else if provider != nil {
			result, err = provider.Complete(r.Context(), request)
		} else {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文本模型执行器未配置，无法生成封面提示词"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "封面提示词生成失败"})
			return
		}
		result = strings.TrimSpace(result)
		if result == "" {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "封面提示词生成失败：模型返回为空"})
			return
		}
		ratio := input.AspectRatio
		if ratio == "" {
			ratio = "3:4"
		}
		if !strings.Contains(result, "--ar") {
			result += "\n\n--ar " + ratio
		}
		writeJSON(w, http.StatusOK, map[string]string{"prompt": result})
	})))
	root.Handle("POST /api/plugins/extract-text", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limitRequestBody(w, r, maxPluginDocumentBodyBytes) {
			return
		}
		if err := r.ParseMultipartForm(maxPluginDocumentBodyBytes); err != nil {
			if isBodyTooLarge(err) {
				writePayloadTooLarge(w, maxPluginDocumentBodyBytes)
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文件读取失败或超过 20MB"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请上传文档文件"})
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, 20<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文件读取失败"})
			return
		}
		name := header.Filename
		var text string
		switch strings.ToLower(filepath.Ext(name)) {
		case ".txt", ".md":
			text = strings.TrimPrefix(string(data), "\ufeff")
		case ".docx":
			text, err = docxPlainText(data)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Word 文档解析失败：" + err.Error()})
				return
			}
		case ".doc":
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "暂不支持 .doc 二进制格式，请在 Word 中另存为 .docx 后上传"})
			return
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 .txt、.doc、.docx"})
			return
		}
		if strings.TrimSpace(text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文档中没有可读取的文字"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"text": text, "name": name})
	})))
	root.Handle("POST /api/plugins/{toolID}/run", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := prompt.PluginToolPromptOptions(r.PathValue("toolID")); !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "插件工具不存在"})
			return
		}
		var input pluginToolRunInput
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		input.Input = strings.TrimSpace(input.Input)
		input.ModelID = strings.TrimSpace(input.ModelID)
		input.PromptID = strings.TrimSpace(input.PromptID)
		if input.Input == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入需要处理的内容"})
			return
		}
		system, ok := prompt.PluginToolSystemPrompt(r.PathValue("toolID"), input.PromptID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择有效的提示词"})
			return
		}
		request := batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: "## 需要处理的内容\n" + input.Input, MaxTokens: 12000, Temperature: 0.45, DisableJSONResponse: true}
		if input.Raw {
			request.UserPrompt = input.Input
		}
		var result string
		var err error
		if execute != nil && input.ModelID != "" {
			identity, _ := BridgeIdentityFromContext(r.Context())
			result, err = execute(r.Context(), identity.Username, input.ModelID, request)
		} else if provider != nil {
			result, err = provider.Complete(r.Context(), request)
		} else {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文本模型执行器未配置"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文本模型请求失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"result": strings.TrimSpace(result)})
	})))
}

// docxPlainText 只提取 Word 正文文字，保留段落换行；图片、页眉页脚和样式忽略。
func docxPlainText(data []byte) (string, error) {
	content, err := openDocxDocumentXML(data)
	if err != nil {
		if errors.Is(err, errDocxMissingDoc) {
			return "", errors.New("缺少 word/document.xml")
		}
		return "", err
	}
	defer content.Close()
	decoder := xml.NewDecoder(content)
	var builder strings.Builder
	inText := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch item := token.(type) {
		case xml.StartElement:
			switch item.Name.Local {
			case "t":
				inText = true
			case "br":
				builder.WriteByte('\n')
			case "tab":
				builder.WriteByte('\t')
			}
		case xml.EndElement:
			switch item.Name.Local {
			case "t":
				inText = false
			case "p":
				builder.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				builder.Write(item)
			}
		}
	}
	return normalizeDocumentText(builder.String()), nil
}

func normalizeDocumentText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}
