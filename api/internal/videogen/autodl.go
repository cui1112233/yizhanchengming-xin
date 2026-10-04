package videogen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AutoDLNoImageWorkflow   = "minimax_h3_lightx2v_no_pic"
	AutoDLReferenceWorkflow = "minimax_h3_lightx2v_v5_15s"
	MaxAutoDLReferences     = 9
)

type AutoDLProvider struct {
	Config      Config
	Client      *http.Client
	ValidateURL URLValidator
}

func (p AutoDLProvider) Submit(ctx context.Context, request Request) (Task, error) {
	cfg, err := NormalizeConfig(p.Config)
	if err != nil { return Task{}, err }
	if ProviderForModel(request.Model, cfg.Provider) != ProviderAutoDLH3 || !ModelMatchesProviderModel(request.Model, cfg.Model, cfg.Provider) {
		return Task{}, errors.New("selected video model does not match AutoDL H3 provider")
	}
	if strings.TrimSpace(request.Prompt)=="" { return Task{}, errors.New("video prompt is required") }
	duration := request.Duration
	if duration <= 0 { duration = 5 }
	if len(request.ReferenceURLs) > MaxAutoDLReferences { return Task{}, fmt.Errorf("AutoDL H3 supports at most %d reference images", MaxAutoDLReferences) }
	resolution := strings.TrimSpace(request.Resolution)
	if resolution=="" { resolution="480p竖" }
	workflow := AutoDLNoImageWorkflow
	payload := map[string]any{"prompt":request.Prompt,"duration":duration,"resolution":resolution}
	if len(request.ReferenceURLs)>0 {
		workflow = AutoDLReferenceWorkflow
		for i, raw := range request.ReferenceURLs {
			raw = strings.TrimSpace(raw)
			if raw=="" { return Task{}, fmt.Errorf("AutoDL H3 reference image %d is empty",i) }
			if err := p.validateMediaURL(raw); err != nil { return Task{}, fmt.Errorf("invalid AutoDL H3 reference image: %w",err) }
			payload[fmt.Sprintf("ref_image_%d",i)] = raw
		}
	}
	endpoint := strings.ReplaceAll(cfg.CreateURL,"{workflow}",workflow)
	reply, err := p.do(ctx,http.MethodPost,endpoint,cfg.APIKey,payload)
	if err != nil { return Task{},err }
	id := firstString(reply,"task_id","taskId","id")
	if id=="" { return Task{},errors.New("AutoDL H3 response did not contain a task id") }
	state := mapAutoDLState(firstString(reply,"status","state"))
	return Task{ID:id,State:state},nil
}

func (p AutoDLProvider) Poll(ctx context.Context, task Task) (Task,error) {
	cfg,err := NormalizeConfig(p.Config)
	if err!=nil { return Task{},err }
	id:=strings.TrimSpace(task.ID)
	if id=="" { return Task{},errors.New("AutoDL H3 task id is required") }
	endpoint:=strings.ReplaceAll(cfg.TasksURL,"{id}",url.PathEscape(id))
	reply,err:=p.do(ctx,http.MethodGet,endpoint,cfg.APIKey,nil)
	if err!=nil { return Task{},err }
	state:=mapAutoDLState(firstString(reply,"status","state"))
	result:=Task{ID:id,State:state}
	if duration:=firstFloat(reply,"actualDurationSeconds","actual_duration_seconds","durationSeconds","duration_seconds"); duration>0 { result.DurationMs=uint64(duration*1000) }
	if state==StateSucceeded {
		mediaURL:=resultMediaURL(reply)
		if mediaURL=="" { return Task{},errors.New("AutoDL H3 completed response did not contain a media URL") }
		if err:=p.validateMediaURL(mediaURL); err!=nil { return Task{},fmt.Errorf("AutoDL H3 returned unsafe media URL: %w",err) }
		result.MediaURL=mediaURL
	}
	return result,nil
}

func mapAutoDLState(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "SUCCESS","SUCCEEDED","COMPLETED","DONE": return StateSucceeded
	case "FAILED","ERROR","CANCELLED","CANCELED": return StateFailed
	case "QUEUED","SUBMITTED","PENDING", "": return StateQueued
	default: return StateRunning
	}
}

func (p AutoDLProvider) do(ctx context.Context,method,endpoint,apiKey string,payload any)([]byte,error){
	if err:=p.validateEndpoint(endpoint); err!=nil { return nil,err }
	var body io.Reader
	if payload!=nil { encoded,err:=json.Marshal(payload); if err!=nil{return nil,err}; body=bytes.NewReader(encoded) }
	req,err:=http.NewRequestWithContext(ctx,method,endpoint,body); if err!=nil{return nil,err}
	req.Header.Set("Accept","application/json")
	req.Header.Set("Authorization",strings.TrimSpace(apiKey))
	if payload!=nil { req.Header.Set("Content-Type","application/json") }
	client:=p.Client; if client==nil { client=&http.Client{Timeout:120*time.Second} }
	resp,err:=client.Do(req); if err!=nil{return nil,errors.New("AutoDL H3 provider request failed")}
	defer resp.Body.Close()
	data,err:=io.ReadAll(io.LimitReader(resp.Body,8<<20)); if err!=nil{return nil,errors.New("read AutoDL H3 provider response")}
	if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("AutoDL H3 provider returned HTTP %d",resp.StatusCode)}
	return data,nil
}

func (p AutoDLProvider) validateEndpoint(raw string) error {
	if p.ValidateURL!=nil{return p.ValidateURL(raw)}
	u,err:=url.Parse(strings.TrimSpace(raw)); if err!=nil||u.Scheme!="https"||strings.ToLower(u.Hostname())!="autodl.art" { return errors.New("AutoDL H3 endpoint must use https://autodl.art") }
	return nil
}
func (p AutoDLProvider) validateMediaURL(raw string) error {
	if p.ValidateURL!=nil{return p.ValidateURL(raw)}
	u,err:=url.Parse(strings.TrimSpace(raw)); if err!=nil||(u.Scheme!="https"&&u.Scheme!="http")||u.Hostname()=="" { return errors.New("invalid AutoDL media URL") }
	return nil
}
