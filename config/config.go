package config

import (
	"fmt"
	"log"
	"time"

	"github.com/NIROOZbx/notification-engine/pkg/jwt"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	HTTPAddr     string        `mapstructure:"http_addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
}

type DatabaseConfig struct {
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MinOpenConns    int           `mapstructure:"min_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	MaxIdleTime     time.Duration `mapstructure:"max_idle_time"`
	DSN             string
}

type RedisConfig struct {
	Addr     string
	Password string
}

type OAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}
type StripeConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	SuccessURL    string `mapstructure:"success_url"`
	CancelURL     string `mapstructure:"cancel_url"`
}

type AuthConfig struct {
	AccessExpiryMinutes int    `mapstructure:"access_expiry_minutes"`
	RefreshExpiryHours  int    `mapstructure:"refresh_expiry_hours"`
	AccessTokenSecret   string `mapstructure:"access_token_secret"`
	RefreshTokenSecret  string `mapstructure:"refresh_token_secret"`
	Environment         string `mapstructure:"environment"`
}

type KafkaConfig struct {
	Broker  string `mapstructure:"broker"`
	GroupID string `mapstructure:"group_id"`
}
type GRPCConfig struct {
	GRPCPort string `mapstructure:"grpc_port"`
}


type Config struct {
	Server    ServerConfig   `mapstructure:"server"`
	Log       LogConfig      `mapstructure:"log"`
	Database  DatabaseConfig `mapstructure:"database"`
	Redis     RedisConfig    `mapstructure:"redis"`
	Auth      AuthConfig     `mapstructure:"auth"`
	OAuth     OAuthConfig    `mapstructure:"oauth"`
	Kafka     KafkaConfig    `mapstructure:"kafka"`
	SecretKey string         `mapstructure:"secret_key"`
	GRPC      GRPCConfig     `mapstructure:"grpc"`
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("no .env file found — reading from environment directly")
	}
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	v.AutomaticEnv()

	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")

	v.BindEnv("redis.password", "REDIS_PASSWORD")

	v.BindEnv("auth.refresh_token_secret", "REFRESH_SECRET")
	v.BindEnv("auth.access_token_secret", "ACCESS_SECRET")

	v.BindEnv("oauth.client_id", "CLIENT_ID")
	v.BindEnv("oauth.client_secret", "CLIENT_SECRET")
	v.BindEnv("oauth.redirect_url", "REDIRECT_URL")

	v.BindEnv("kafka.broker", "KAFKA_BROKER")
	v.BindEnv("kafka.group_id", "KAFKA_GROUP_ID")

	v.BindEnv("secret_key", "CREDENTIALS_SECRET")


	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.Database.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	validate(&cfg)
	InitOAuth(&cfg.OAuth)
	log.Println("Loaded all configs ✅")
	return &cfg, nil

}

func validate(cfg *Config) {
	rules := []struct {
		value  string
		envVar string
	}{
		{cfg.Database.User, "DB_USER"},
		{cfg.Database.Password, "DB_PASSWORD"},
		{cfg.Database.Host, "DB_HOST"},
		{cfg.Database.Name, "DB_NAME"},
		{cfg.Redis.Addr, "REDIS_ADDR"},
		{cfg.Auth.AccessTokenSecret, "ACCESS_SECRET"},
		{cfg.Auth.RefreshTokenSecret, "REFRESH_SECRET"},
		{cfg.OAuth.ClientID, "CLIENT_ID"},
		{cfg.OAuth.ClientSecret, "CLIENT_SECRET"},
		{cfg.OAuth.RedirectURL, "REDIRECT_URL"},
		{cfg.Kafka.Broker, "KAFKA_BROKER"},
		{cfg.SecretKey, "CREDENTIALS_SECRET"},
	}

	for _, rule := range rules {
		if rule.value == "" {
			log.Fatalf("%s is required", rule.envVar)
		}
	}
}

func InitOAuth(cfg *OAuthConfig) {
	goth.UseProviders(google.New(
		cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, "email", "profile",
	))

	log.Println("OAuth providers initialized successfully ✅")

}

func (a *AuthConfig) ToJWTConfig() jwt.Config {
	return jwt.Config{
		AccessTokenSecret:   a.AccessTokenSecret,
		RefreshTokenSecret:  a.RefreshTokenSecret,
		AccessExpiryMinutes: a.AccessExpiryMinutes,
		RefreshExpiryHours:  a.RefreshExpiryHours,
	}
}
