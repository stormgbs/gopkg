package http_reverse_proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

type EndpointAddr struct {
	IP   string
	Port int
}

func (ea *EndpointAddr) String() string {
	return fmt.Sprintf("%s:%d", ea.IP, ea.Port)
}

var trans = &http.Transport{
	DisableKeepAlives:   false,
	MaxIdleConnsPerHost: 300,
	DisableCompression:  true,
}

type ProxyEndpointFunc func() EndpointAddr

func NewProxyRoundTripper(pefnc ProxyEndpointFunc) *proxyRoundTripper {
	return &proxyRoundTripper{
		proxyEndpointFunc: pefnc,
		transport:         trans,
	}
}

type proxyRoundTripper struct {
	proxyEndpointFunc ProxyEndpointFunc
	transport         http.RoundTripper
}

func (rrt *proxyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	addr := rrt.proxyEndpointFunc()
	r.URL.Host = addr.String()
	return rrt.transport.RoundTrip(r)
}

func NewReverseProxy(proxyTransport http.RoundTripper, r *http.Request) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = r.Host
			r.URL.Opaque = r.RequestURI
			r.URL.RawQuery = ""
		},
		Transport:     proxyTransport,
		FlushInterval: 50 * time.Millisecond,
	}
}

////////////////////////////////   demo  test   //////////////////////////////////////////////////////
var redirect_http_methods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
}

//直接作为返回HandlerFunc
func MakeReverseProxyHandler(r *http.Request) http.HandlerFunc {
	rt := NewProxyRoundTripper(MyTestProxyEndpointFunc)
	return NewReverseProxy(rt, r).ServeHTTP
}

func MyTestProxyEndpointFunc() EndpointAddr {
	return EndpointAddr{
		IP:   "127.0.0.1",
		Port: 7788,
	}
}

func MyTestHttpHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if _, ok := redirect_http_methods[r.Method]; ok {
		rt := NewProxyRoundTripper(MyTestProxyEndpointFunc)
		NewReverseProxy(rt, r).ServeHTTP(w, r)
		return
	}

	//本机处理请求
	// ...
	return
}
