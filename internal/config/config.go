package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Node     NodeConfig     `mapstructure:"node"`
	API      HttpConfig     `mapstructure:"api"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type NodeConfig struct {
	Mode string `mapstructure:"mode"`
}

type HttpConfig struct {
	Port string `mapstructure:"port"`
	URL  string `mapstructure:"url"`
}

type GRPCConfig struct {
	Port string `mapstructure:"port"`
	URL  string `mapstructure:"url"`
}

type LoggerConfig struct {
	BufferSize     int         `mapstructure:"buffer_size"`
	WorkerCount    int         `mapstructure:"worker_count"`
	IsDev          bool        `mapstructure:"is_dev"`
	AddSource      bool        `mapstructure:"add_source"`
	BaseAttributes []slog.Attr `mapstructure:"base_attributes"`
	Dir            string      `mapstructure:"dir"`
	FileName       string      `mapstructure:"fileName"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	SSLMode         string        `mapstructure:"sslmode"`
	Debug           bool          `mapstructure:"debug"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DefaultTTL   int    `mapstructure:"default_ttl"`
	OrderTTL     int    `mapstructure:"order_ttl"`
	Enabled      bool   `mapstructure:"enabled"`
}

type ConfigOption struct {
	Key          string
	DefaultValue interface{}
	EnvKey       string
}

type ConfigLoader struct {
	v       *viper.Viper
	options []ConfigOption
}

func NewConfigLoader() *ConfigLoader {
	cl := &ConfigLoader{
		v:       viper.New(),
		options: make([]ConfigOption, 0),
	}
	cl.initializeOptions()
	return cl
}

func (cl *ConfigLoader) initializeOptions() {
	cl.options = []ConfigOption{
		// Node options
		{Key: "node.mode", EnvKey: "NODE_MODE", DefaultValue: "development"},

		// Http options
		{Key: "http.port", EnvKey: "API_PORT", DefaultValue: "8080"},
		{Key: "http.url", EnvKey: "API_URL", DefaultValue: "localhost"},

		// GRPC options
		{Key: "grpc.port", EnvKey: "GRPC_PORT", DefaultValue: "50051"},
		{Key: "grpc.url", EnvKey: "GRPC_URL", DefaultValue: "localhost"},

		// Logger
		{Key: "logger.dir", EnvKey: "", DefaultValue: "logs"},
		{Key: "logger.filename", EnvKey: "", DefaultValue: "app.log"},

		// Postgres options
		{Key: "postgres.host", EnvKey: "POSTGRES_HOST", DefaultValue: "localhost"},
		{Key: "postgres.port", EnvKey: "POSTGRES_PORT", DefaultValue: "5432"},
		{Key: "postgres.user", EnvKey: "POSTGRES_USER", DefaultValue: "postgres"},
		{Key: "postgres.password", EnvKey: "POSTGRES_PASSWORD", DefaultValue: "postgres"},
		{Key: "postgres.dbname", EnvKey: "POSTGRES_DBNAME", DefaultValue: "myapp"},
		{Key: "postgres.max_idle_conns", EnvKey: "POSTGRES_MAX_IDLE_CONNS", DefaultValue: 10},
		{Key: "postgres.max_open_conns", EnvKey: "POSTGRES_MAX_OPEN_CONNS", DefaultValue: 100},
		{Key: "postgres.conn_max_lifetime", EnvKey: "POSTGRES_CONN_MAX_LIFETIME", DefaultValue: time.Hour},
		{Key: "postgres.conn_max_idle_time", EnvKey: "POSTGRES_CONN_MAX_IDLE_TIME", DefaultValue: time.Minute * 30},
		{Key: "postgres.retry_attempts", EnvKey: "POSTGRES_RETRY_ATTEMPTS", DefaultValue: 3},
		{Key: "postgres.retry_delay", EnvKey: "POSTGRES_RETRY_DELAY", DefaultValue: time.Second * 5},
		{Key: "postgres.sslmode", EnvKey: "POSTGRES_SSLMODE", DefaultValue: "disable"},
		{Key: "postgres.debug", EnvKey: "POSTGRES_DEBUG", DefaultValue: false},

		// Redis options
		{Key: "redis.host", EnvKey: "REDIS_HOST", DefaultValue: "localhost"},
		{Key: "redis.port", EnvKey: "REDIS_PORT", DefaultValue: "6379"},
		{Key: "redis.password", EnvKey: "REDIS_PASSWORD", DefaultValue: ""},
		{Key: "redis.db", EnvKey: "REDIS_DB", DefaultValue: 0},
		{Key: "redis.pool_size", EnvKey: "REDIS_POOL_SIZE", DefaultValue: 10},
		{Key: "redis.min_idle_conns", EnvKey: "REDIS_MIN_IDLE_CONNS", DefaultValue: 5},
		{Key: "redis.default_ttl", EnvKey: "REDIS_DEFAULT_TTL", DefaultValue: 3600},
		{Key: "redis.enabled", EnvKey: "REDIS_ENABLED", DefaultValue: true},
	}
}

func LoadConfig() (*Config, error) {
	loader := NewConfigLoader()
	return loader.Load()
}

func (cl *ConfigLoader) Load() (*Config, error) {
	if err := cl.setDefaults(); err != nil {
		return nil, fmt.Errorf("error setting defaults: %w", err)
	}

	if err := cl.loadConfigFile(); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	var config Config
	if err := cl.v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func (cl *ConfigLoader) setDefaults() error {
	for _, opt := range cl.options {
		value := opt.DefaultValue
		if opt.EnvKey != "" {
			if envValue, exists := os.LookupEnv(opt.EnvKey); exists {
				value = envValue
			}
		}
		cl.v.SetDefault(opt.Key, value)
	}
	return nil
}

func (cl *ConfigLoader) loadConfigFile() error {
	cl.v.SetConfigName("config")
	cl.v.SetConfigType("yaml")
	cl.v.AddConfigPath(".")
	cl.v.AddConfigPath("./config")
	cl.v.AddConfigPath("../config")
	cl.v.AutomaticEnv()

	if err := cl.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found. Creating default config file...")
			return cl.createDefaultConfigFile()
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	log.Printf("Using config file: %s", cl.v.ConfigFileUsed())
	return nil
}

func (cl *ConfigLoader) createDefaultConfigFile() error {
	if err := cl.v.SafeWriteConfig(); err != nil {
		return fmt.Errorf("error writing default config file: %w", err)
	}
	log.Println("Default config file created. Please edit it with your settings and restart the application.")
	return nil
}