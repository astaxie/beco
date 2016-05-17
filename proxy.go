package main

import (
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
)

const (
	DefaultTimeout = time.Second * 30
	DefaultWeight  = 1
)

type BackendList []*BackendProxy

func NewBackendList(backends ...Backend) (BackendList, int, error) {
	var (
		bl        BackendList
		maxWeight int
	)
	for _, backend := range backends {
		u, err := url.Parse(backend.Host)
		if err != nil {
			return bl, maxWeight, err
		}
		if backend.Weight == 0 {
			backend.Weight = DefaultWeight
		}
		if backend.Weight > maxWeight {
			maxWeight = backend.Weight
		}
		rp := httputil.NewSingleHostReverseProxy(u)
		if backend.FailTimeout == 0 {
			backend.FailTimeout = DefaultTimeout
		}
		rp.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   backend.FailTimeout,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		bl = append(bl, &BackendProxy{
			target: u,
			server: rp,
		})
	}
	return bl, maxWeight, nil
}

// Proxy is a http proxy capable of also proxying any connection, includeing websockets.
type BackendProxy struct {
	target *url.URL
	server *httputil.ReverseProxy
}

func (b *BackendProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.target.User != nil {
		credentials := []byte(b.target.User.String())
		encoded := base64.StdEncoding.EncodeToString(credentials)
		r.Header.Set("Authorization", "Basic "+encoded)
	}
	if !isWebsocket(r) {
		b.server.ServeHTTP(w, r)
		return
	}
	target := *b.target
	target.Path = r.URL.Path
	proxy := b.websocket(&target)
	proxy.ServeHTTP(w, r)
	return
}

func (b *BackendProxy) websocket(target *url.URL) *websocketproxy.WebsocketProxy {
	wsTarget := *target
	wsTarget.Scheme = "ws"
	if target.Scheme == "https" {
		wsTarget.Scheme = "wss"
	}
	proxy := websocketproxy.NewProxy(&wsTarget)
	proxy.Upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return proxy
}

// isWebsocket checks wether the incoming request is a part of websocket
// handshake
func isWebsocket(req *http.Request) bool {
	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" ||
		!strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
		return false
	}
	return true
}

func ProxyHandler(proxy Proxy) (http.Handler, error) {
	backendList, maxWeight, err := NewBackendList(proxy.Backends...)
	i := -1
	gcd := 1
	cw := 0
	// Weighted Round-Robin Scheduling
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, header := range proxy.SetHeaders {
			w.Header().Set(header.Key, header.Value)
		}
		for {
			i = (i + 1) % len(proxy.Backends)
			if i == 0 {
				cw = cw - gcd
				if cw <= 0 {
					cw = maxWeight
				}
			}
			if proxy.Backends[i].Weight >= cw {
				backendList[i].ServeHTTP(w, r)
				break
			}
		}
	}), err
}
