package operator

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all operator configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Runtime  RuntimeConfig  `mapstructure:"runtime"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	GRPCAddr        string `mapstructure:"grpc_addr"`
	OperatorAddress string `mapstructure:"operator_address"`
}

type RuntimeConfig struct {
	Type       string           `mapstructure:"type"`
	Docker     DockerConfig     `mapstructure:"docker"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
}

type DockerConfig struct {
	Network string          `mapstructure:"network"`
	Host    string          `mapstructure:"host"` // Remote Docker daemon URL (e.g. "tcp://192.168.1.100:2376"), empty for local socket
	TLS     DockerTLSConfig `mapstructure:"tls"`
}

type DockerTLSConfig struct {
	CACert string `mapstructure:"ca_cert"` // Path to CA certificate
	Cert   string `mapstructure:"cert"`    // Path to client certificate
	Key    string `mapstructure:"key"`     // Path to client key
}

type KubernetesConfig struct {
	Namespace       string `mapstructure:"namespace"`
	WorkerNamespace string `mapstructure:"worker_namespace"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

// LoadConfig reads configuration from file and environment variables.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.grpc_addr", ":50051")
	v.SetDefault("server.operator_address", "portwhine-operator:50051")
	v.SetDefault("runtime.type", "docker")
	v.SetDefault("runtime.docker.network", "portwhine")
	v.SetDefault("runtime.kubernetes.namespace", "portwhine")
	v.SetDefault("runtime.kubernetes.worker_namespace", "portwhine-workers")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "portwhine")
	v.SetDefault("database.user", "portwhine")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "168h")
	v.SetDefault("log.level", "info")

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("operator")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("/etc/portwhine")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Environment variables with PW_ prefix
	v.SetEnvPrefix("PW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind all config keys to env vars so that Unmarshal picks them up.
	// Viper's AutomaticEnv only works with Get() calls, not with Unmarshal.
	keys := []string{
		"server.grpc_addr",
		"server.operator_address",
		"runtime.type",
		"runtime.docker.network",
		"runtime.docker.host",
		"runtime.docker.tls.ca_cert",
		"runtime.docker.tls.cert",
		"runtime.docker.tls.key",
		"runtime.kubernetes.namespace",
		"runtime.kubernetes.worker_namespace",
		"database.host",
		"database.port",
		"database.name",
		"database.user",
		"database.password",
		"database.sslmode",
		"auth.jwt_secret",
		"auth.access_token_ttl",
		"auth.refresh_token_ttl",
		"log.level",
	}
	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("bind env %s: %w", key, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
