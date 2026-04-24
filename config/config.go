package configs

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/spf13/viper"
)

type conf struct {
	DBDriver           string   `mapstructure:"DB_DRIVER"`
	DBHost             string   `mapstructure:"DB_HOST"`
	DBPort             string   `mapstructure:"DB_PORT"`
	DBUser             string   `mapstructure:"DB_USER"`
	DBPassword         string   `mapstructure:"DB_PASSWORD"`
	DBName             string   `mapstructure:"DB_NAME"`
	WebServerPort      string   `mapstructure:"WEB_SERVER_PORT"`
	GRPCServerPort     string   `mapstructure:"GRPC_SERVER_PORT"`
	InitialChatMessage string   `mapstructure:"INITIAL_CHAT_MESSAGE"`
	OpenAIApiKey       string   `mapstructure:"OPENAI_API_KEY"`
	Model              string   `mapstructure:"MODEL"`
	ModelMaxTokens     int      `mapstructure:"MODEL_MAX_TOKENS"`
	Temperature        float32  `mapstructure:"TEMPERATURE"`
	TopP               float32  `mapstructure:"TOP_P"`
	N                  int      `mapstructure:"N"`
	Stop               []string
	MaxTokens          int    `mapstructure:"MAX_TOKENS"`
	AuthToken          string `mapstructure:"AUTH_TOKEN"`
}

func LoadConfig(path string) (*conf, error) {
	var cfg conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(filepath.Join(path, ".env"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	stop, err := parseStop(viper.GetString("STOP"))
	if err != nil {
		return nil, err
	}
	cfg.Stop = stop

	return &cfg, nil
}

func parseStop(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}

	var stop []string
	if err := json.Unmarshal([]byte(raw), &stop); err != nil {
		return nil, errors.New("invalid STOP config: expected JSON array of strings")
	}

	if stop == nil {
		return []string{}, nil
	}

	return stop, nil
}
