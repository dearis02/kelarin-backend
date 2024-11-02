package config

import (
	"fmt"
	pkg "kelarin/pkg/validator"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type PostgresConfig struct {
	ConString   string `yaml:"connection_string"`
	MaxIdleCons int    `yaml:"max_idle_cons"`
	MaxOpenCons int    `yaml:"max_open_cons"`
}

func (p PostgresConfig) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.ConString, validation.Required),
		validation.Field(&p.MaxIdleCons, validation.Required),
		validation.Field(&p.MaxOpenCons, validation.Required),
	)
}

type RedisConfig struct {
	ConString string `yaml:"connection_string"`
}

func (r RedisConfig) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ConString, validation.Required),
	)
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"cors"`
	MaxRequest uint `yaml:"max_request"`
}

func (s Server) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.Port, validation.Required),
		validation.Field(&s.CORS, validation.Required),
	)
}

type JWTConfig struct {
	SecretKey              string        `yaml:"secret_key"`
	Issuer                 string        `yaml:"issuer"`
	Expiration             time.Duration `yaml:"expiration"`
	RefreshTokenSecretKey  string        `yaml:"refresh_token_secret_key"`
	RefreshTokenExpiration time.Duration `yaml:"refresh_token_expiration"`
}

func (j JWTConfig) Validate() error {
	return validation.ValidateStruct(&j,
		validation.Field(&j.SecretKey, validation.Required, validation.By(pkg.ValidateBase64)),
		validation.Field(&j.Issuer, validation.Required),
		validation.Field(&j.Expiration, validation.Required),
		validation.Field(&j.RefreshTokenSecretKey, validation.Required, validation.By(pkg.ValidateBase64)),
		validation.Field(&j.RefreshTokenExpiration, validation.Required),
	)
}

type OAuthConfig struct {
	Google GoogleOAuth `yaml:"google"`
}

func (o OAuthConfig) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Google, validation.Required),
	)
}

type GoogleOAuth struct {
	ClientId string `yaml:"client_id"`
}

func (g GoogleOAuth) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.ClientId, validation.Required),
	)
}

type Config struct {
	Environment string         `yaml:"environment"`
	Server      Server         `yaml:"server"`
	DataBase    PostgresConfig `yaml:"database"`
	Redis       RedisConfig    `yaml:"redis"`
	JWT         JWTConfig      `yaml:"jwt"`
	PrettyLog   bool           `yaml:"pretty_log"`
	Oauth       OAuthConfig    `yaml:"oauth"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Server, validation.Required),
		validation.Field(&c.DataBase, validation.Required),
		validation.Field(&c.Redis, validation.Required),
		validation.Field(&c.JWT, validation.Required),
		validation.Field(&c.Oauth, validation.Required),
	)
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) Mode() string {
	if c.Environment == "production" {
		return gin.ReleaseMode
	} else {
		return gin.DebugMode
	}
}

func NewAppConfig() *Config {
	cfg := &Config{}

	cfgData, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config file")
	}

	err = yaml.Unmarshal(cfgData, &cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config file")
	}

	err = cfg.Validate()
	if err != nil {
		log.Fatal().Err(err).Msg("Config validation failed")
	}

	return cfg
}
