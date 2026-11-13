package safehttp

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"
)

var outboundDecorator struct {
	sync.RWMutex
	wrap    func(http.RoundTripper) http.RoundTripper
	managed func(http.RoundTripper, time.Duration) http.RoundTripper
}

// SetManagedOutboundDecorator moves request timeout ownership inside admission,
// so time spent waiting for capacity cannot consume a paid provider deadline.
func SetManagedOutboundDecorator(wrap func(http.RoundTripper, time.Duration) http.RoundTripper) func() {
	outboundDecorator.Lock()
	previous := outboundDecorator.managed
	outboundDecorator.managed = wrap
	outboundDecorator.Unlock()
	return func() { outboundDecorator.Lock(); outboundDecorator.managed = previous; outboundDecorator.Unlock() }
}
func managedOutbound(base http.RoundTripper, timeout time.Duration) (http.RoundTripper, time.Duration) {
	outboundDecorator.RLock()
	wrap := outboundDecorator.managed
	outboundDecorator.RUnlock()
	if wrap != nil {
		return wrap(base, timeout), 0
	}
	return decorateOutbound(base), timeout
}

// SetOutboundDecorator installs application-wide admission control without
// replacing the underlying SSRF checks. Call once before serving requests.
// The restoration closure is for isolated tests and application shutdown.
func SetOutboundDecorator(wrap func(http.RoundTripper) http.RoundTripper) func() {
	outboundDecorator.Lock()
	previous := outboundDecorator.wrap
	outboundDecorator.wrap = wrap
	outboundDecorator.Unlock()
	return func() { outboundDecorator.Lock(); outboundDecorator.wrap = previous; outboundDecorator.Unlock() }
}

func decorateOutbound(base http.RoundTripper) http.RoundTripper {
	outboundDecorator.RLock()
	wrap := outboundDecorator.wrap
	outboundDecorator.RUnlock()
	if wrap != nil {
		return wrap(base)
	}
	return base
}

// Options 描述一个出站客户端。所有字段都可留空，留空时使用默认值。
type Options struct {
	// Timeout 是单次请求（含重定向与读完响应体）的总超时。
	// 调用方原有的业务超时应写入这里，统一客户端不会用默认值覆盖它。
	Timeout time.Duration
	// ConnectTimeout 是 TCP 建连超时。
	ConnectTimeout time.Duration
	// ResponseHeaderTimeout 是等待响应头的超时；0 表示沿用 Transport 自身配置。
	ResponseHeaderTimeout time.Duration
	// MaxRedirects 是允许的最大重定向跳数，0 表示使用 DefaultMaxRedirects。
	MaxRedirects int
	// Transport 允许调用方保留专用 Transport（例如出图链路强制 HTTP/1.1 或
	// HTTP/2、自定义 Proxy）。为 *http.Transport 时会被克隆并追加拨号校验。
	Transport http.RoundTripper
}

// NewClient 构造一个带 SSRF 防护与逐跳重定向校验的客户端。
func NewClient(options Options) *http.Client {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = envDuration(envTimeout)
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	redirects := options.MaxRedirects
	if redirects <= 0 {
		redirects = DefaultMaxRedirects
	}
	transport, timeout := managedOutbound(newTransport(options), timeout)
	return &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		CheckRedirect: redirectPolicy(redirects),
	}
}

// NewClientWithTimeout 是只关心总超时的便捷构造。
func NewClientWithTimeout(timeout time.Duration) *http.Client {
	return NewClient(Options{Timeout: timeout})
}

// NewClientWithTransport 保留调用方的专用 Transport，并追加统一防护。
func NewClientWithTransport(timeout time.Duration, transport *http.Transport) *http.Client {
	return NewClient(Options{Timeout: timeout, Transport: transport})
}

// NewClientWithTransportWithoutTimeout 保留统一的拨号防护和重定向检查，
// 但不设置整个请求的绝对总超时。适用于响应头已有独立超时、正文
// 由调用方按「无数据进展」检测断流的大文件下载。
func NewClientWithTransportWithoutTimeout(transport *http.Transport) *http.Client {
	outbound, _ := managedOutbound(newTransport(Options{Transport: transport}), 0)
	return &http.Client{
		Transport:     outbound,
		CheckRedirect: redirectPolicy(DefaultMaxRedirects),
	}
}

// NewTransport 只返回带防护的 Transport，供必须沿用“无整体超时、由调用方
// context 控制”的 Client（例如大文件上传）使用。
func NewTransport(options Options) http.RoundTripper {
	transport, _ := managedOutbound(newTransport(options), options.Timeout)
	return transport
}

func newTransport(options Options) http.RoundTripper {
	base, ok := options.Transport.(*http.Transport)
	if ok && base == nil {
		ok = false
	}
	keepConcreteTransport := ok
	if !ok {
		if options.Transport != nil {
			// 非 *http.Transport 无法安全克隆并替换拨号器，至少保留请求级校验。
			return &guardRoundTripper{base: options.Transport}
		}
		base = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		base = base.Clone()
	}
	connectTimeout := options.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = envDuration(envConnectTimeout)
	}
	if connectTimeout <= 0 {
		connectTimeout = DefaultConnectTimeout
	}
	proxyAllowList := proxyDialAllowList(base)
	base.DialContext = (&net.Dialer{
		Timeout:   connectTimeout,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			if allowPrivateDestinations.Load() {
				return nil
			}
			// 配置里显式放行的地址（如豆包图片中继）也要在拨号层识别，
			// 否则请求级放行后连接仍会被拒；匹配要求 IP 与端口同时一致。
			if isExplicitlyAllowedDial(address) {
				return nil
			}
			if _, allowed := proxyAllowList[address]; allowed {
				return nil
			}
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return fmt.Errorf("%w：无法识别目标地址", ErrBlockedAddress)
			}
			ip := net.ParseIP(host)
			if ip == nil || IsBlockedIP(ip) {
				return fmt.Errorf("%w（%s）", ErrBlockedAddress, host)
			}
			return nil
		},
	}).DialContext
	if options.ResponseHeaderTimeout > 0 {
		base.ResponseHeaderTimeout = options.ResponseHeaderTimeout
	}
	if keepConcreteTransport {
		// 调用方传入的是专用 *http.Transport（强制 HTTP/1.1、HTTP/2、专用 Proxy），
		// 保持其具体类型，避免破坏既有的传输能力检测。安全约束仍然生效：
		// 拨号阶段由 Control 拒绝内网地址，重定向由 CheckRedirect 逐跳校验。
		return base
	}
	return &guardRoundTripper{base: base}
}

// proxyDialAllowList 返回允许拨号的代理地址。显式代理（如 TTS 的
// 127.0.0.1:7890）本身是回环地址，不能因为拨号校验被误拦；目标地址在默认
// 传输下由 guardRoundTripper 请求级校验，专用传输下由拨号校验与重定向
// 逐跳校验兜底。
func proxyDialAllowList(base *http.Transport) map[string]struct{} {
	allow := map[string]struct{}{}
	if base == nil || base.Proxy == nil {
		return allow
	}
	probe, err := base.Proxy(&http.Request{URL: &url.URL{Scheme: "https", Host: "example.invalid"}})
	if err != nil || probe == nil {
		return allow
	}
	host := probe.Hostname()
	if host == "" {
		return allow
	}
	port := probe.Port()
	if port == "" {
		if probe.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	add := func(ip string) {
		allow[net.JoinHostPort(ip, port)] = struct{}{}
	}
	add(host)
	if net.ParseIP(host) == nil {
		if addresses, lookupErr := net.LookupIP(host); lookupErr == nil {
			for _, address := range addresses {
				add(address.String())
			}
		}
	}
	return allow
}

type guardRoundTripper struct {
	base http.RoundTripper
}

func (transport *guardRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request == nil || request.URL == nil {
		return nil, fmt.Errorf("%w：地址为空", ErrBlockedAddress)
	}
	if err := CheckURL(request.URL.String()); err != nil {
		return nil, err
	}
	return transport.base.RoundTrip(request)
}

func redirectPolicy(maxRedirects int) func(*http.Request, []*http.Request) error {
	return func(request *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("%w（最多 %d 跳）", ErrTooManyRedirects, maxRedirects)
		}
		return CheckURL(request.URL.String())
	}
}
