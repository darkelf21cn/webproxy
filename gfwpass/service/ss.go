package service

import (
	"context"
	"fmt"
	"gfwpass/util"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const SSBin = "/usr/bin/ss-local"

type ShadowSocksProxy struct {
	ctx        context.Context
	cancel     context.CancelFunc
	remoteHost string
	remotePort string
	localPort  int
	cipher     string
	password   string
	plugin     string
	pluginOpts string
	latency    time.Duration
}

func (s ShadowSocksProxy) Executable() string {
	return SSBin
}

func (s *ShadowSocksProxy) Ping(timeout time.Duration) {
	lat, _ := ping(s.remoteHost, s.remotePort, timeout)
	s.latency = lat
}

func (s ShadowSocksProxy) Latency() time.Duration {
	return s.latency
}

func (s *ShadowSocksProxy) Start() (func(), error) {
	cmd := exec.CommandContext(s.ctx, s.Executable(), s.parseArgs()...)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	stopFunc := func() {
		s.cancel()
		cmd.Wait()
	}
	return stopFunc, nil
}

func (s *ShadowSocksProxy) String() string {
	return fmt.Sprintf("%s:%s", s.remoteHost, s.remotePort)
}

func (s *ShadowSocksProxy) Cmd() string {
	cmd := append([]string{s.Executable()}, s.parseArgs()...)
	return strings.Join(cmd, " ")
}

func (s ShadowSocksProxy) parseArgs() []string {
	// ss-local

	// -s <server_host>           Host name or IP address of your remote server.
	// -p <server_port>           Port number of your remote server.
	// -l <local_port>            Port number of your local server.
	// -k <password>              Password of your remote server.
	// -m <encrypt_method>        Encrypt method: rc4-md5,
	// 						   aes-128-gcm, aes-192-gcm, aes-256-gcm,
	// 						   aes-128-cfb, aes-192-cfb, aes-256-cfb,
	// 						   aes-128-ctr, aes-192-ctr, aes-256-ctr,
	// 						   camellia-128-cfb, camellia-192-cfb,
	// 						   camellia-256-cfb, bf-cfb,
	// 						   chacha20-ietf-poly1305,
	// 						   xchacha20-ietf-poly1305,
	// 						   salsa20, chacha20 and chacha20-ietf.
	// 						   The default cipher is chacha20-ietf-poly1305.

	// [-a <user>]                Run as another user.
	// [-f <pid_file>]            The file path to store pid.
	// [-t <timeout>]             Socket timeout in seconds.
	// [-c <config_file>]         The path to config file.
	// [-n <number>]              Max number of open files.
	// [-i <interface>]           Network interface to bind.
	// [-b <local_address>]       Local address to bind.

	// [-u]                       Enable UDP relay.
	// [-U]                       Enable UDP relay and disable TCP relay.

	// [--reuse-port]             Enable port reuse.
	// [--fast-open]              Enable TCP fast open.
	// 						   with Linux kernel > 3.7.0.
	// [--acl <acl_file>]         Path to ACL (Access Control List).
	// [--mtu <MTU>]              MTU of your network interface.
	// [--mptcp]                  Enable Multipath TCP on MPTCP Kernel.
	// [--no-delay]               Enable TCP_NODELAY.
	// [--key <key_in_base64>]    Key of your remote server.
	// [--plugin <name>]          Enable SIP003 plugin. (Experimental)
	// [--plugin-opts <options>]  Set SIP003 plugin options. (Experimental)

	// [-v]                       Verbose mode.
	// [-h, --help]               Print this message.
	args := []string{
		"-s", s.remoteHost,
		"-p", s.remotePort,
		"-l", fmt.Sprintf("%d", s.localPort),
		"-m", s.cipher,
		"-k", s.password,
	}
	if s.plugin != "" {
		args = append(args, "--plugin", s.plugin)
	}
	if s.pluginOpts != "" {
		args = append(args, "--plugin-opts", s.pluginOpts)
	}
	return args
}

func NewShadowSocksProxy(ctx context.Context, u *url.URL, localPort int) (IProxy, error) {
	c, cancel := context.WithCancel(ctx)
	p := &ShadowSocksProxy{
		ctx:        c,
		cancel:     cancel,
		remoteHost: u.Hostname(),
		remotePort: u.Port(),
		localPort:  localPort,
	}
	if u.User != nil {
		_, pwdSet := u.User.Password()
		if !pwdSet {
			tmp, err := util.Base64Decode(u.User.Username())
			if err != nil {
				return nil, err
			}
			i := strings.IndexRune(string(tmp), ':')
			if i >= 0 {
				p.cipher = string(tmp)[:i]
				p.password = string(tmp)[i+1:]
			}
		} else {
			p.cipher = u.User.Username()
			p.password, _ = u.User.Password()
		}
	}
	tmp := u.Query().Get("plugin")
	if tmp != "" {
		i := strings.IndexRune(tmp, ';')
		if i >= 0 {
			p.plugin = tmp[:i]
			p.pluginOpts = tmp[i+1:]
		}
	}
	return p, nil
}
