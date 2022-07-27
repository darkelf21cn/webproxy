package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	ENV_LOGLEVEL      = "GFWPASS_LOGLEVEL"
	ENV_PORT          = "GFWPASS_PORT"
	ENV_SUBS_URL      = "GFWPASS_SUBS_URL"
	ENV_SUBS_INTERVAL = "GFWPASS_SUBS_INTERVAL_HOUR"
	ENV_HC_URLS       = "GFWPASS_HC_URLS"
	ENV_HC_INTERVAL   = "GFWPASS_HC_INTERVAL_SEC"
	ENV_HC_TIMEOUT    = "GFWPASS_HC_TIMEOUT_SEC"
	ENV_HC_APPEMPTS   = "GFWPASS_HC_APPEMPTS"
)

type Config struct {
	LogLevel                        string      `yaml:"LogLevel"`
	Port                            int         `yaml:"Port"`
	SubscriptionURL                 string      `yaml:"SubscribeURL"`
	SubscriptionUpdateIntervalHours int         `yaml:"SubscriptionUpdateIntervalHours"`
	HealthCheck                     HealthCheck `yaml:"HealthCheck"`
}

type HealthCheck struct {
	URLs        []string `yaml:"URLs"`
	IntervalSec int64    `yaml:"IntervalSec"`
	TimeoutSec  int      `yaml:"TimeoutSec"`
	Attempts    int      `yaml:"Attempts"`
}

func newDefaultConfig() Config {
	return Config{
		LogLevel:                        "info",
		SubscriptionUpdateIntervalHours: 24,
		Port:                            1080,
		HealthCheck: HealthCheck{
			URLs: []string{
				"https://www.google.com",
			},
			IntervalSec: 60,
			TimeoutSec:  5,
			Attempts:    3,
		},
	}
}

func LoadConfig(file string) Config {
	if file == "" {
		return loadConfigFromEnv()
	} else {
		return loadConfigFromFile(file)
	}
}

func loadConfigFromFile(file string) Config {
	config := newDefaultConfig()

	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to read config: %s", err.Error()))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("failed to marshal config: %s", err.Error()))
	}
	return config
}

func loadConfigFromEnv() Config {
	config := newDefaultConfig()
	if v := os.Getenv(ENV_LOGLEVEL); v != "" {
		config.LogLevel = v
	}
	if v := os.Getenv(ENV_PORT); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			config.Port = int(i)
		}
	}
	if v := os.Getenv(ENV_SUBS_URL); v != "" {
		config.SubscriptionURL = v
	}
	if v := os.Getenv(ENV_SUBS_INTERVAL); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			config.SubscriptionUpdateIntervalHours = int(i)
		}
	}
	if v := os.Getenv(ENV_HC_URLS); v != "" {
		urls := strings.Split(v, ",")
		for i := 0; i < len(urls); i++ {
			urls[i] = strings.TrimSpace(urls[i])
		}
		config.HealthCheck.URLs = urls
	}
	if v := os.Getenv(ENV_HC_INTERVAL); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			config.HealthCheck.IntervalSec = i
		}
	}
	if v := os.Getenv(ENV_HC_TIMEOUT); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			config.HealthCheck.TimeoutSec = int(i)
		}
	}
	if v := os.Getenv(ENV_HC_APPEMPTS); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			config.HealthCheck.Attempts = int(i)
		}
	}
	return config
}

func (a *Config) String() string {
	b, err := yaml.Marshal(a)
	if err != nil {
		return "null"
	}
	return string(b)
}
