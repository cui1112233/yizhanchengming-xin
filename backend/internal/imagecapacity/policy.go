package imagecapacity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

type Request struct {
	Endpoint, Credential, Model, Size, Quality, Resolution, Owner, Book string
	References                                                          bool
	Count                                                               int
}

type Policy struct {
	configEpoch     string
	AccountKey      string `json:"accountKey"`
	GroupKey        string `json:"groupKey"`
	ProfileKey      string `json:"profileKey"`
	Model           string `json:"model"`
	Size            string `json:"size"`
	Quality         string `json:"quality"`
	Resolution      string `json:"resolution"`
	References      bool   `json:"references"`
	Count           int    `json:"count"`
	AccountLimit    int    `json:"accountLimit"`
	ModelLimit      int    `json:"modelLimit"`
	ModelTotalLimit int    `json:"modelTotalLimit"`
	InitialLimit    int    `json:"initialLimit"`
	Verified        bool   `json:"verified"`
}

type Repository interface {
	Resolve(context.Context, Request) (Policy, error)
	Save(context.Context, Policy) error
}

func hash(v any) string {
	b, _ := json.Marshal(v)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func AccountKey(r Request) string {
	u, err := url.Parse(strings.TrimSpace(r.Endpoint))
	host := strings.ToLower(strings.TrimSpace(r.Endpoint))
	if err == nil && u.Host != "" {
		host = strings.TrimRight(strings.ToLower(u.Hostname()), ".")
		port := u.Port()
		if port != "" && !((strings.EqualFold(u.Scheme, "https") && port == "443") || (strings.EqualFold(u.Scheme, "http") && port == "80")) {
			host = net.JoinHostPort(host, port)
		}
	}
	return hash([]string{host, strings.TrimSpace(r.Credential)})
}
func ProfileKey(r Request) string {
	return hash([]any{modelIdentity(r.Model), strings.TrimSpace(r.Size), r.References, strings.TrimSpace(r.Quality), strings.TrimSpace(r.Resolution), imageCount(r.Count)})
}
func modelIdentity(model string) string { return strings.ToLower(strings.TrimSpace(model)) }
func imageCount(n int) int {
	if n <= 0 {
		return 1
	}
	return n
}
func defaultPolicy(r Request) Policy {
	return Policy{AccountKey: AccountKey(r), ProfileKey: ProfileKey(r), Model: r.Model, Size: r.Size, Quality: r.Quality, Resolution: r.Resolution, References: r.References, Count: imageCount(r.Count), AccountLimit: 4, ModelLimit: 4, ModelTotalLimit: 4, InitialLimit: 4}
}
func normalize(p Policy, r Request) Policy {
	p.AccountKey = AccountKey(r)
	p.ProfileKey = ProfileKey(r)
	p.Count = imageCount(r.Count)
	if p.AccountLimit < 1 {
		p.AccountLimit = 4
	}
	if p.ModelLimit < 1 {
		p.ModelLimit = 4
	}
	if p.ModelTotalLimit < 1 {
		p.ModelTotalLimit = p.ModelLimit
	}
	if p.InitialLimit < 1 {
		p.InitialLimit = 4
	}
	if p.InitialLimit > p.ModelLimit {
		p.InitialLimit = p.ModelLimit
	}
	if p.InitialLimit < p.Count && p.Count <= p.ModelLimit {
		p.InitialLimit = p.Count
	}
	return p
}
