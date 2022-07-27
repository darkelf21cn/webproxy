package service

import (
	"context"
	"fmt"
	"gfwpass/conf"
	"gfwpass/util"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type IProxy interface {
	Executable() string
	Ping(timeout time.Duration)
	Latency() time.Duration
	Start() error
	Stop()
	String() string
	Cmd() string
}

type Proxies []IProxy

func (s Proxies) Len() int {
	return len(s)
}

func (s Proxies) Less(i, j int) bool {
	return s[i].Latency() < s[j].Latency()
}

func (s Proxies) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Service struct {
	logger  *zap.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	proxies Proxies

	healthCheckURLs            []string
	healthCheckInterval        time.Duration
	healthCheckTimeout         time.Duration
	healthCheckAttempts        int
	subscriptionURL            string
	subscriptionUpdateInterval time.Duration
	port                       int
}

func NewService(logger *zap.Logger, conf conf.Config) (*Service, error) {
	s := &Service{
		logger:                     logger,
		healthCheckURLs:            conf.HealthCheck.URLs,
		healthCheckInterval:        time.Duration(conf.HealthCheck.IntervalSec * int64(time.Second)),
		healthCheckTimeout:         time.Duration(int64(conf.HealthCheck.TimeoutSec) * int64(time.Second)),
		healthCheckAttempts:        conf.HealthCheck.Attempts,
		subscriptionURL:            conf.SubscriptionURL,
		subscriptionUpdateInterval: time.Duration(int64(conf.SubscriptionUpdateIntervalHours) * int64(time.Hour)),
		port:                       conf.Port,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	err := s.subscribeProxyServers()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Start() error {
	if len(s.proxies) == 0 {
		return fmt.Errorf("no server available")
	}

	go s.RunSubscriptionReloaderDaemon(s.ctx)
	s.logger.Info("subscription reloader daemon started")
	for {
		tryNextProxy := func() {
			// stop the server, put it to the end of the queue
			time.Sleep(2 * time.Second)
			s.proxies = append(s.proxies[1:], s.proxies[0])
			s.logger.Info("try next proxy")
		}

		// start the first server
		s.logger.Info("starting proxy", zap.String("cmd", s.proxies[0].Cmd()))
		err := s.proxies[0].Start()
		if err != nil {
			s.logger.Info("failed to start the proxy", zap.Error(err))
			tryNextProxy()
			continue
		}
		s.logger.Info("proxy started")

		// delay health-check
		time.Sleep(5 * time.Second)

		// start health-check
		code := s.RunHealthCheckDaemon(s.ctx)
		if code < 0 {
			return nil
		} else {
			s.logger.Info("stopping proxy due to health-check failed")
			s.proxies[0].Stop()
			waitPortToBeFree(s.logger, s.port)
			tryNextProxy()
			continue
		}
	}
}

func (s *Service) Stop() error {
	if s.cancel == nil {
		s.logger.Info("service has stopped")
		return nil
	}
	s.cancel()
	s.logger.Debug("stopping service")
	s.cancel = nil
	return nil
}

func (s *Service) RunHealthCheckDaemon(ctx context.Context) int {
	s.logger.Info("starting health-check daemon")
	client, err := util.NewHttpClient(fmt.Sprintf(":%d", s.port), s.healthCheckTimeout)
	if err != nil {
		s.logger.Error("failed to start the proxy, try the next one", zap.Error(err))
		return -1
	}

	// Fast fail on startup
	if err := s.checkHealth(client, 1); err != nil {
		s.logger.Info(err.Error())
		return 0
	}

	// Regular checks
	s.logger.Info("health-check daemon started", zap.String("interval", s.healthCheckInterval.String()))
	ticker := time.NewTicker(s.healthCheckInterval)
	for {
		select {
		case <-ctx.Done():
			return -2
		case <-ticker.C:
			if err := s.checkHealth(client, s.healthCheckAttempts); err != nil {
				s.logger.Info(err.Error())
				return 1
			}
		}
	}
}

func (s *Service) RunSubscriptionReloaderDaemon(ctx context.Context) {
	s.logger.Info("starting subscription reloader daemon")
	ticker := time.NewTicker(s.subscriptionUpdateInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.subscribeProxyServers()
		}
	}
}

func (s *Service) checkHealth(client *http.Client, attempts int) error {
	urls := append(make([]string, 0), s.healthCheckURLs...)
	for i := 1; i <= attempts; i++ {
		ok := true
		for j := 0; j < len(urls); j++ {
			url := urls[j]
			_, err := client.Get(url)
			if err != nil {
				ok = false
				s.logger.Info(fmt.Sprintf("probe %s failed at %d/%d attempt", url, i, attempts))
			} else {
				s.logger.Debug(fmt.Sprintf("probe %s ok at %d/%d attempt", url, i, attempts))
				urls = append(urls[:j], urls[j+1:]...)
				j -= 1
			}
		}
		if ok {
			s.logger.Info("health-check ok")
			return nil
		}
		if i < attempts {
			time.Sleep(s.healthCheckInterval)
		}
	}
	return fmt.Errorf("health-check failed after %d attempts", attempts)
}

func (s *Service) subscribeProxyServers() error {
	// Get proxy servers from the subscription URL
	client, err := util.NewHttpClient("", 30*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.Get(s.subscriptionURL)
	if err != nil {
		return err
	}
	base64Raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	raw, err := util.Base64Decode(string(base64Raw))
	if err != nil {
		return err
	}
	urls := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")

	// Parse proxy server URL
	executables := make(map[string]struct{}, 0)
	proxies := make(Proxies, 0)
	for _, urlStr := range urls {
		u, err := url.Parse(urlStr)
		if err != nil {
			s.logger.Info(fmt.Sprintf("invalid proxy url, skipping %s", urlStr))
			continue
		}
		var proxy IProxy
		switch strings.ToLower(u.Scheme) {
		case "ss":
			var err error
			proxy, err = NewShadowSocksProxy(u, s.port)
			if err != nil {
				s.logger.Error(fmt.Sprintf("invalid shadowsocks proxy url, skipping %s", urlStr), zap.Error(err))
				continue
			}
		case "":
			continue
		default:
			s.logger.Info(fmt.Sprintf("unknown proxy protocol, skipping %s", urlStr))
			continue
		}
		s.logger.Debug("proxy url parsed", zap.String("proxy", proxy.String()))
		proxies = append(proxies, proxy)
		executables[proxy.Executable()] = struct{}{}
	}
	for executable := range executables {
		if _, err := os.Stat(executable); err != nil {
			return err
		}
	}
	proxies = testProxyLatency(s.logger, proxies, 2*time.Second)
	s.logger.Info("printing available proxies")
	for _, proxy := range proxies {
		s.logger.Info(fmt.Sprintf("%s %dms", proxy.String(), proxy.Latency().Milliseconds()))
	}
	s.proxies = proxies
	return nil
}

func ping(host, port string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return 9223372036854775807, err
	}
	defer conn.Close()
	return time.Since(start), nil
}

func waitPortToBeFree(logger *zap.Logger, port int) {
	logger.Info(fmt.Sprintf("wait until port %d is free", port))
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			logger.Debug("port is busy", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		} else {
			ln.Close()
			break
		}
	}
	logger.Info(fmt.Sprintf("port %d is free", port))
}

func testProxyLatency(logger *zap.Logger, proxies Proxies, timeout time.Duration) Proxies {
	logger.Info("testing proxy network latency")
	wg := &sync.WaitGroup{}
	for _, svc := range proxies {
		logger.Debug(svc.String())
		wg.Add(1)
		go func(wg *sync.WaitGroup, svc IProxy) {
			defer wg.Done()
			svc.Ping(timeout)
		}(wg, svc)
	}
	wg.Wait()
	sort.Sort(proxies)
	for len(proxies) > 0 {
		last := len(proxies) - 1
		if proxies[last].Latency() > timeout {
			logger.Debug(fmt.Sprintf("excluding proxy: %s", proxies[last].String()))
			proxies = proxies[:last]
		} else {
			return proxies
		}
	}
	return nil
}
