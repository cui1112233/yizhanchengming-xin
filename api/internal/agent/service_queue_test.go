package agent

import (
	"context"
	"encoding/json"
	"testing"
)

type toolQueueFake struct { ids []string; err error }
func (q *toolQueueFake) Enqueue(_ context.Context, id string) error { q.ids = append(q.ids, id); return q.err }

func TestServiceQueuesCreativeToolButNotNavigation(t *testing.T) {
	store := &memoryStore{thread: Thread{ID:"thread-1",Owner:"owner",Title:"测试",Status:"active"}}
	videoArgs, _ := json.Marshal(map[string]any{"model":"yd2-mini-video","prompt":"镜头推进","duration":5})
	queue := &toolQueueFake{}
	service := &Service{Store:store,Responder:&fixedResponder{response:AgentResponse{Content:"开始生成视频",Tool:&ToolProposal{Name:ToolVideoGenerate,Label:"生成视频",Arguments:videoArgs}}},ToolQueue:queue}
	result, err := service.SendMessage(context.Background(),"owner","thread-1",SendMessageInput{Content:"生成视频"})
	if err != nil { t.Fatal(err) }
	if result.ToolCall == nil || len(queue.ids) != 1 || queue.ids[0] != result.ToolCall.ID { t.Fatalf("queue=%#v tool=%#v",queue.ids,result.ToolCall) }

	store2 := &memoryStore{thread: Thread{ID:"thread-2",Owner:"owner",Title:"测试",Status:"active"}}
	navArgs, _ := json.Marshal(NavigateArgs{Path:"/novel-fetch"})
	queue2 := &toolQueueFake{}
	service2 := &Service{Store:store2,Responder:&fixedResponder{response:AgentResponse{Content:"打开",Tool:&ToolProposal{Name:ToolNavigate,Label:"打开",Arguments:navArgs}}},ToolQueue:queue2}
	if _, err := service2.SendMessage(context.Background(),"owner","thread-2",SendMessageInput{Content:"打开小说获取"}); err != nil { t.Fatal(err) }
	if len(queue2.ids) != 0 { t.Fatalf("navigation must remain browser-executed: %#v",queue2.ids) }
}
