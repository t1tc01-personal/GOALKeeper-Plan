package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const configFile = "config.yaml"

// Configurations - Configuration for the application
type Configurations struct {
	AppConfig      AppConfig      `mapstructure:"app"`
	PostgresConfig PostgresConfig `mapstructure:"postgres"`
	AuthConfig     AuthConfig     `mapstructure:"auth"`
	RedisConfig    RedisConfig    `mapstructure:"redis"`
}

type AppConfig struct {
	Name        string     `mapstructure:"name"`
	Environment string     `mapstructure:"environment"`
	Version     string     `mapstructure:"version"`
	ServiceName string     `mapstructure:"serviceName"`
	ServerName  string     `mapstructure:"serverName"`
	Debug       bool       `mapstructure:"debug"`
	Timezone    string     `mapstructure:"timezone"`
	Port        string     `mapstructure:"port"`
	CORS        CORSConfig `mapstructure:"cors"`
}

// CORSConfig - Configuration for CORS
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	ExposedHeaders   []string `mapstructure:"exposedHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"` // in seconds
}

// PoolConfig - Configuration for pool
type PoolConfig struct {
	MaxOpenConnections    int `mapstructure:"maxOpenConnections"`
	MaxIdleConnections    int `mapstructure:"maxIdleConnections"`
	MaxConnectionLifetime int `mapstructure:"maxConnectionLifetime"`
	MaxConnectionIdleTime int `mapstructure:"maxConnectionIdleTime"`
}

// PostgresConfig - Configuration for postgres
type PostgresConfig struct {
	PoolConfig       `mapstructure:",squash"`
	ConnectionString string `mapstructure:"connectionString"`
}

// AuthConfig - Configuration for Auth
type AuthConfig struct {
	Secret            string `mapstructure:"secret"`
	JWTExpiration     int    `mapstructure:"jwtExpiration"`     // in hours
	RefreshExpiration int    `mapstructure:"refreshExpiration"` // in days
	GitHubClientID    string `mapstructure:"githubClientId"`
	GitHubSecret      string `mapstructure:"githubSecret"`
	GoogleClientID    string `mapstructure:"googleClientId"`
	GoogleSecret      string `mapstructure:"googleSecret"`
	FrontendURL       string `mapstructure:"frontendUrl"`
	BackendURL        string `mapstructure:"backendUrl"`
}

type RedisConfig struct {
	RedisHost     string `mapstructure:"redisHost"`
	RedisPort     string `mapstructure:"redisPort"`
	RedisPassword string `mapstructure:"redisPassword"`
}

// InitConfig initializes the config.
func InitConfig() *Configurations {
	viper.SetConfigType("yaml")

	// Expand environment variables inside the config file
	b, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("read config has failed with error: %v\n", err)
		os.Exit(1)
	}

	expand := os.ExpandEnv(string(b))
	configReader := strings.NewReader(expand)

	viper.AutomaticEnv()

	if err := viper.ReadConfig(configReader); err != nil {
		fmt.Printf("read config has failed with error: %v\n", err)
		os.Exit(1)
	}

	configs := Configurations{}
	if err := viper.Unmarshal(&configs); err != nil {
		fmt.Printf("read config has failed with error: %v\n", err)
		os.Exit(1)
	}

	return &configs
}

var lock = &sync.Mutex{}
var singleConfig *Configurations

func GetConfig() *Configurations {
	if singleConfig == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleConfig == nil {
			singleConfig = InitConfig()
		}
	}

	return singleConfig
}
