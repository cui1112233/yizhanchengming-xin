// Package safehttp 提供统一的出站 HTTP 客户端。
//
// 后端访问外部地址（文本模型、图片模型、配音、网页抓取、下载远端图片等）
// 统一走这里的客户端，获得三项约束：
//   - 拒绝内网、回环、链路本地、云元数据等保留地址（含域名解析结果）；
//   - 重定向逐跳重新校验，跳数有上限；
//   - 连接、响应头、整体超时统一可配，调用方仍可覆盖自己的总超时。
//
// 内网服务（本地执行器、4001 健康检查等）不属于 SSRF 面，继续使用各自的
// 客户端，不接入本包。
package safehttp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultTimeout 是未显式指定总超时时的默认值。生成类接口耗时较长，
	// 调用方通常会用 Options.Timeout 覆盖成自己的业务超时。
	DefaultTimeout = 120 * time.Second
	// DefaultConnectTimeout 与 net.DefaultTransport 的拨号超时保持一致，
	// 避免给图片等慢链路引入额外回归。
	DefaultConnectTimeout = 30 * time.Second
	// DefaultMaxRedirects 是允许跟随的最大重定向跳数。
	DefaultMaxRedirects = 5
)

const (
	envTimeout        = "QIANTIE_OUTBOUND_TIMEOUT"
	envConnectTimeout = "QIANTIE_OUTBOUND_CONNECT_TIMEOUT"
	// envDoubaoRelayBaseURL 是豆包图片中继的地址。该中继是桌面客户端里的
	// Tauri 本地服务，按设计只能监听本机回环，改配置改不掉。这里只放行
	// 配置里写明的这一个「协议 + 主机 + 端口」，不做网段放行；未配置该
	// 变量时，127.0.0.1 仍然全部拦截。
	envDoubaoRelayBaseURL = "QIANTIE_DOUBAO_RELAY_BASE_URL"
	// envOutboundAllowURLs 是额外放行的出站地址列表，逗号或空白分隔，例如
	// `QIANTIE_OUTBOUND_ALLOW_URLS=http://127.0.0.1:18900`。用于本机受信任
	// 的模型网关等回环服务；每项仍然要求协议、主机、端口三者完全一致，
	// 不做网段或通配符放行。
	envOutboundAllowURLs = "QIANTIE_OUTBOUND_ALLOW_URLS"
)

var (
	// ErrBlockedAddress 表示目标解析到内网/回环/保留地址。
	ErrBlockedAddress = errors.New("出站请求被拒绝：目标地址属于内网、回环或保留地址")
	// ErrBlockedScheme 表示协议不是 http/https。
	ErrBlockedScheme = errors.New("出站请求被拒绝：只允许 http/https 协议")
	// ErrTooManyRedirects 表示重定向次数超过上限。
	ErrTooManyRedirects = errors.New("出站请求被拒绝：重定向次数超过上限")
	// ErrResolveFailed 表示目标域名无法解析，无法判断是否安全。
	ErrResolveFailed = errors.New("出站请求被拒绝：无法解析目标主机")
)

// allowPrivateDestinations 仅供测试二进制使用。httptest 监听 127.0.0.1，
// 如果不放行，所有使用本包的既有测试都会被误杀。生产二进制
// （qiantie-local.exe / qiantie-executor.exe 等）不满足 .test 后缀，
// 因此线上始终拦截内网地址。
var allowPrivateDestinations = func() *atomic.Bool {
	value := &atomic.Bool{}
	name := strings.ToLower(filepath.Base(os.Args[0]))
	if strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe") {
		value.Store(true)
	}
	return value
}()

// IsBlockedIP 判断 IP 是否落在禁止出站的地址段。
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		ip = v4
		if v4.IsLoopback() || v4.IsPrivate() || v4.IsLinkLocalUnicast() ||
			v4.IsLinkLocalMulticast() || v4.IsUnspecified() || v4.IsMulticast() {
			return true
		}
		// 0.0.0.0/8：部分系统会把 0.x 路由到本机。
		if v4[0] == 0 {
			return true
		}
		// 100.64.0.0/10：运营商级 NAT，阿里云元数据 100.100.100.200 落在其中。
		if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return true
		}
		// 198.18.0.0/15：基准测试保留网段。
		if v4[0] == 198 && (v4[1] == 18 || v4[1] == 19) {
			return true
		}
		// 192.0.0.0/24：IETF 协议分配网段。
		if v4[0] == 192 && v4[1] == 0 && v4[2] == 0 {
			return true
		}
		// 240.0.0.0/4：保留网段（含广播地址 255.255.255.255）。
		if v4[0] >= 240 {
			return true
		}
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	return false
}

// explicitAllowEntry 是一条显式放行记录。只有协议、主机、端口三者同时一致
// 才算命中；域名形式还会解析后逐个地址核对，不会因为域名相同就跳过 IP 黑名单。
type explicitAllowEntry struct {
	source string
	scheme string
	host   string
	port   string
	raw    string
	ips    []net.IP
}

var (
	allowMu       sync.Mutex
	allowCacheRaw string
	allowCache    []explicitAllowEntry
	allowCacheAt  time.Time
	allowLogged   = map[string]struct{}{}
	// allowCacheTTL 控制域名放行记录的重新解析间隔，避免每次拨号都做 DNS。
	allowCacheTTL = 30 * time.Second
)

// explicitAllowEntries 返回当前配置允许放行的地址。未配置、解析失败或协议
// 不受支持时都返回空列表，保持“默认不放行任何内网地址”。首次解析成功时
// 打印一行日志，说明放行了哪个地址、来自哪个配置项。
func explicitAllowEntries() []explicitAllowEntry {
	rawDoubao := strings.TrimSpace(os.Getenv(envDoubaoRelayBaseURL))
	rawList := strings.TrimSpace(os.Getenv(envOutboundAllowURLs))
	cacheKey := rawDoubao + "\n" + rawList
	if strings.TrimSpace(cacheKey) == "" {
		return nil
	}
	allowMu.Lock()
	defer allowMu.Unlock()
	if allowCacheRaw == cacheKey && time.Since(allowCacheAt) < allowCacheTTL {
		return allowCache
	}
	allowCacheRaw = cacheKey
	allowCacheAt = time.Now()
	entries := []explicitAllowEntry{}
	if rawDoubao != "" {
		entries = append(entries, resolveExplicitAllow(envDoubaoRelayBaseURL, rawDoubao)...)
	}
	for _, raw := range strings.FieldsFunc(rawList, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	}) {
		entries = append(entries, resolveExplicitAllow(envOutboundAllowURLs, raw)...)
	}
	allowCache = entries
	for _, entry := range allowCache {
		key := entry.source + "|" + entry.raw
		if _, logged := allowLogged[key]; logged {
			continue
		}
		allowLogged[key] = struct{}{}
		log.Printf("[safehttp] 已显式放行地址 %s（来自配置项 %s）", entry.raw, entry.source)
	}
	return allowCache
}

func resolveExplicitAllow(source, raw string) []explicitAllowEntry {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil
	}
	host := parsed.Hostname()
	if host == "" {
		return nil
	}
	entry := explicitAllowEntry{
		source: source,
		scheme: scheme,
		host:   host,
		port:   effectivePort(parsed),
		raw:    scheme + "://" + parsed.Host,
	}
	if ip := net.ParseIP(host); ip != nil {
		entry.ips = []net.IP{ip}
		return []explicitAllowEntry{entry}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return nil
	}
	for _, address := range addresses {
		entry.ips = append(entry.ips, address.IP)
	}
	return []explicitAllowEntry{entry}
}

// matchesRequest 要求协议、主机、端口完全一致；域名形式会重新解析并逐个
// 地址核对，新增的解析结果不在放行集合内时仍然拒绝。
func (entry explicitAllowEntry) matchesRequest(parsed *url.URL) bool {
	if parsed == nil || !strings.EqualFold(parsed.Scheme, entry.scheme) {
		return false
	}
	if !strings.EqualFold(parsed.Hostname(), entry.host) {
		return false
	}
	if effectivePort(parsed) != entry.port {
		return false
	}
	if net.ParseIP(entry.host) != nil {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, entry.host)
	if err != nil || len(addresses) == 0 {
		return false
	}
	for _, address := range addresses {
		if !entry.containsIP(address.IP) {
			return false
		}
	}
	return true
}

// matchesDial 只接受放行集合里明确列出的 IP 与配置端口的组合，同机其它
// 端口或其它地址仍然会被拨号校验拒绝。
func (entry explicitAllowEntry) matchesDial(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != entry.port {
		return false
	}
	return entry.containsIP(net.ParseIP(strings.Trim(host, "[]")))
}

func (entry explicitAllowEntry) containsIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, allowed := range entry.ips {
		if allowed.Equal(ip) {
			return true
		}
	}
	return false
}

func effectivePort(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	if port := parsed.Port(); port != "" {
		return port
	}
	if strings.EqualFold(parsed.Scheme, "https") {
		return "443"
	}
	return "80"
}

func isExplicitlyAllowedURL(parsed *url.URL) bool {
	for _, entry := range explicitAllowEntries() {
		if entry.matchesRequest(parsed) {
			return true
		}
	}
	return false
}

func isExplicitlyAllowedDial(address string) bool {
	for _, entry := range explicitAllowEntries() {
		if entry.matchesDial(address) {
			return true
		}
	}
	return false
}

// CheckURL 校验一个出站地址：协议必须是 http/https，主机必须解析到公网地址。
// 域名会实际解析一次；只要任一解析结果命中保留地址就拒绝。
func CheckURL(raw string) error {
	return checkURL(raw, allowPrivateDestinations.Load())
}

func checkURL(raw string, allowPrivate bool) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("%w：地址格式无效", ErrBlockedScheme)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return ErrBlockedScheme
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("%w：缺少主机名", ErrBlockedAddress)
	}
	if allowPrivate {
		return nil
	}
	// 配置里明确指定的那一个地址可以放行，其余内网地址仍然拦截。
	if isExplicitlyAllowedURL(parsed) {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		if IsBlockedIP(ip) {
			return fmt.Errorf("%w（%s）", ErrBlockedAddress, host)
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("%w（%s）", ErrResolveFailed, host)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("%w（%s）", ErrResolveFailed, host)
	}
	for _, address := range addresses {
		if IsBlockedIP(address.IP) {
			return fmt.Errorf("%w（%s -> %s）", ErrBlockedAddress, host, address.IP)
		}
	}
	return nil
}

func envDuration(key string) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}
