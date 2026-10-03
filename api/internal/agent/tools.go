package agent

import (
	"encoding/json"
	"strings"
)

const ToolNavigate = "ui.navigate"

type NavigateArgs struct {
	Path    string `json:"path"`
	Section string `json:"section,omitempty"`
}

type ToolProposal struct {
	Name      string          `json:"name"`
	Label     string          `json:"label"`
	Arguments json.RawMessage `json:"arguments"`
}

type TaskProposal struct {
	Title  string `json:"title"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type AgentResponse struct {
	Content string        `json:"content"`
	Tool    *ToolProposal `json:"tool,omitempty"`
	Task    *TaskProposal `json:"task,omitempty"`
}

type Destination struct {
	Name    string
	Path    string
	Aliases []string
}

var destinations = []Destination{
	{Name: "小说获取", Path: "/novel-fetch", Aliases: []string{"小说获取", "获取小说"}},
	{Name: "剧本生成", Path: "/script", Aliases: []string{"剧本生成", "剧本工作台"}},
	{Name: "创作漫剧", Path: "/shuihuo-production/creative", Aliases: []string{"创作漫剧", "漫剧创作"}},
	{Name: "批量工厂", Path: "/batch-factory", Aliases: []string{"批量工厂", "batch factory"}},
	{Name: "水货生产", Path: "/shuihuo-production", Aliases: []string{"水货生产"}},
	{Name: "API 配置", Path: "/api-config", Aliases: []string{"api配置", "api 配置", "模型配置", "模型设置"}},
	{Name: "设置", Path: "/settings", Aliases: []string{"系统设置", "设置"}},
	{Name: "Agent 工作区", Path: "/agent", Aliases: []string{"agent工作区", "agent 工作区", "agent"}},
}

func destinationForText(text string) (Destination, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	for _, destination := range destinations {
		for _, alias := range destination.Aliases {
			if strings.Contains(text, strings.ToLower(alias)) {
				return destination, true
			}
	}
	return Destination{}, false
}

func navigationTool(destination Destination, section string) *ToolProposal {
	body, _ := json.Marshal(NavigateArgs{Path: destination.Path, Section: section})
	label := "打开「" + destination.Name + "」"
	return &ToolProposal{Name: ToolNavigate, Label: label, Arguments: body}
}

func IsAllowedNavigation(path string) bool {
	for _, destination := range destinations {
		if destination.Path == path {
			return true
		}
	}
	return false
}
