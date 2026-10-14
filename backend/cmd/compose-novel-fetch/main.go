package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	loginPath    = "/batch-rewrite/121-login-hotfix.js"
	feedbackPath = "/batch-rewrite/interaction-feedback.js"
	taskPath     = "/batch-rewrite/task-visibility-hotfix.js"
	runtimePath  = "/batch-rewrite/v78-novel-fetch-v2-runtime.js"
	appPath      = "/batch-rewrite/app.js"
	selectedPath = "/batch-rewrite/v78-selected-task-submit-bridge.js"
	clientPath   = "/batch-rewrite/v78-novel-fetch-v2.js"
	datePath     = "/batch-rewrite/v78-novel-fetch-v2-date.js"
	configPath   = "/batch-rewrite/v78-novel-fetch-v2-config.js"
	layoutPath   = "/batch-rewrite/v78-novel-fetch-v2-layout.js"
	quickPath    = "/batch-rewrite/v78-quick-submit-selection-hotfix.js"
	guardPath    = "/batch-rewrite/v78-selected-submit-payload-guard.js"
	controlsPath = "/batch-rewrite/v78-novel-fetch-v2-run-controls.js"
	versionPath  = "/batch-rewrite/version-config-authority.js"
	loginTag     = `<script id="qiantie-121-login-hotfix" src="/batch-rewrite/121-login-hotfix.js?v=20260915-session-probe-r2"></script>`
	feedbackTag  = `<script id="qiantie-novel-fetch-interaction-feedback" src="/batch-rewrite/interaction-feedback.js?v=20260915-parent-status-r1"></script>`
	taskTag      = `<script id="qiantie-novel-fetch-task-visibility" src="/batch-rewrite/task-visibility-hotfix.js?v=20260906-task-visibility1"></script>`
	runtimeTag   = `<script id="qiantie-novel-fetch-runtime" src="/batch-rewrite/v78-novel-fetch-v2-runtime.js?v=20260916-performance-r1"></script>`
	appTag       = `<script src="/batch-rewrite/app.js?v=20260922-retry-progress-r1"></script>`
	selectedTag  = `<script src="/batch-rewrite/v78-selected-task-submit-bridge.js?v=20260917-task-list-fallback-r1"></script>`
	clientTag    = `<script src="/batch-rewrite/v78-novel-fetch-v2.js?v=20260925-processing-flow-r1"></script>`
	dateTag      = `<script src="/batch-rewrite/v78-novel-fetch-v2-date.js?v=20260922-rendered-date-r1"></script>`
	configTag    = `<script src="/batch-rewrite/v78-novel-fetch-v2-config.js?v=20260916-performance-r1"></script>`
	layoutTag    = `<script src="/batch-rewrite/v78-novel-fetch-v2-layout.js?v=20261013-submit-feedback-r3"></script>`
	quickTag     = `<script src="/batch-rewrite/v78-quick-submit-selection-hotfix.js?v=20260917-capture-before-tab-r1"></script>`
	guardTag     = `<script src="/batch-rewrite/v78-selected-submit-payload-guard.js?v=20260917-payload-guard-r1"></script>`
	controlsTag  = `<script src="/batch-rewrite/v78-novel-fetch-v2-run-controls.js?v=20260925-text-model-layout-r2"></script>`
	versionTag   = `<script src="/batch-rewrite/version-config-authority.js?v=20260909-single-authority-r1"></script>`
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "用法：compose-novel-fetch 源文件 目标文件")
		os.Exit(2)
	}
	source, err := os.ReadFile(filepath.Clean(os.Args[1]))
	if err != nil {
		panic(err)
	}
	composed := inject(string(source))
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(os.Args[2])), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Clean(os.Args[2]), []byte(composed), 0o644); err != nil {
		panic(err)
	}
}

func inject(source string) string {
	if !strings.Contains(strings.ToLower(source), "</body>") {
		return source
	}
	for _, path := range []string{loginPath, feedbackPath, taskPath, runtimePath, selectedPath, clientPath, datePath, configPath, layoutPath, quickPath, guardPath, controlsPath, versionPath} {
		source = dedupeScriptPath(source, path)
	}
	loginRE := scriptForPath(loginPath)
	source = loginRE.ReplaceAllString(source, loginTag)
	appRE := regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=["'][^"']*/app\.js(?:\?[^"']*)?["'][^>]*></script>`)
	runtimeInjected := strings.Contains(source, runtimePath)
	if !runtimeInjected {
		if app := appRE.FindString(source); app != "" {
			source = strings.Replace(source, app, runtimeTag+"\n    "+app, 1)
			runtimeInjected = true
		}
	}
	if app := appRE.FindString(source); app != "" {
		source = strings.Replace(source, app, appTag, 1)
	}
	add := make([]string, 0, 13)
	for _, item := range []struct{ path, tag string }{{loginPath, loginTag}, {feedbackPath, feedbackTag}, {taskPath, taskTag}, {selectedPath, selectedTag}, {clientPath, clientTag}, {datePath, dateTag}, {configPath, configTag}, {layoutPath, layoutTag}, {quickPath, quickTag}, {guardPath, guardTag}, {controlsPath, controlsTag}, {versionPath, versionTag}} {
		if !strings.Contains(source, item.path) {
			add = append(add, item.tag)
		}
	}
	if !runtimeInjected {
		add = append([]string{runtimeTag}, add...)
	}
	if len(add) == 0 {
		return source
	}
	return strings.Replace(source, "</body>", "    "+strings.Join(add, "\n    ")+"\n  </body>", 1)
}

func scriptForPath(path string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(path)
	return regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=["'][^"']*` + escaped + `(?:\?[^"']*)?["'][^>]*></script>`)
}

func dedupeScriptPath(source, path string) string {
	re := scriptForPath(path)
	kept := false
	return re.ReplaceAllStringFunc(source, func(tag string) string {
		if kept {
			return ""
		}
		kept = true
		return tag
	})
}
