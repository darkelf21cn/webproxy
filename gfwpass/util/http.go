package util

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"golang.org/x/net/proxy"
)

func NewHttpClient(proxyAddr string, timeout time.Duration) (*http.Client, error) {
	baseDialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	var dailer proxy.ContextDialer
	if proxyAddr != "" {
		proxyDialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, baseDialer)
		if err != nil {
			return nil, err
		}
		if dialer, ok := proxyDialer.(proxy.ContextDialer); ok {
			dailer = dialer
		} else {
			return nil, fmt.Errorf("failed to create socks5 proxy dialer")
		}
	} else {
		dailer = baseDialer
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dailer.DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}, nil
}
