package main

import (
	"errors"
	"fmt"
	api "goalkeeper-plan/cmd/api"
	"goalkeeper-plan/config"
	"goalkeeper-plan/internal/logger"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const configFile = "config.yaml"

// run server with CLI
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "GOALKeeper-Plan API server CLI",
	Long:  "Run GOALKeeper-Plan API server with CLI",
}

// init initializes the env and logger.
func init() {
	initEnv()
	configs := initConfig()
	logger := initLogger(configs)

	apiCmd := api.NewServerCmd(configs, logger)
	rootCmd.AddCommand(apiCmd)
}

// main is the entry point for the run command.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("run command has failed with error: %v\n", err)
		os.Exit(1)
	}
}

// initEnv loads the .env file
func initEnv() {
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("no .env file found, skipping")
		return
	}
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("init env has failed with error: %v\n", err)
		os.Exit(1)
	}
}

// initLogger creates a new zap Logger
func initLogger(config config.Configurations) logger.Logger {
	logger, err := logger.NewLogger(config.AppConfig)
	if err != nil {
		fmt.Printf("init logger has failed with error: %v\n", err)
		os.Exit(1)
	}
	return logger
}

// initConfig initializes the config.
func initConfig() config.Configurations {
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

	configs := config.Configurations{}
	if err := viper.Unmarshal(&configs); err != nil {
		fmt.Printf("read config has failed with error: %v\n", err)
		os.Exit(1)
	}

	return configs
}
