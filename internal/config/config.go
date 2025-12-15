package config

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/rbnis/handlich/internal/logger"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	LogLevel slog.Level
	Backend  struct {
		Type BackendType `yaml:"type" envconfig:"BACKEND_TYPE"`
		File struct {
			Mode FileBackendMode `yaml:"mode" envconfig:"BACKEND_FILE_MODE"`
			Path string          `yaml:"path" envconfig:"BACKEND_FILE_PATH"`
		} `yaml:"file"`
		Redis struct {
			Host     string `yaml:"host" envconfig:"BACKEND_REDIS_HOST"`
			Username string `yaml:"username" envconfig:"BACKEND_REDIS_USERNAME"`
			Password string `yaml:"password" envconfig:"BACKEND_REDIS_PASSWORD"`
		} `yaml:"redis"`
	} `yaml:"backend"`
}

type BackendType string

const (
	BackendTypeMemory BackendType = "memory"
	BackendTypeFile   BackendType = "file"
	BackendTypeRedis  BackendType = "redis"
	BackendTypeSqlite BackendType = "sqlite"
)

type FileBackendMode string

const (
	FileBackendModeRead  FileBackendMode = "read-only"
	FileBackendModeWrite FileBackendMode = "read-write"
)

func (c *Config) setDefaults() {
	if c.LogLevel == 0 {
		c.LogLevel = slog.LevelInfo
	}
	if c.Backend.Type == "" {
		c.Backend.Type = BackendTypeMemory
	}
	if c.Backend.Type == BackendTypeFile {
		if c.Backend.File.Mode == "" {
			c.Backend.File.Mode = FileBackendModeRead
		}
	}
}

func (c *Config) validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(BackendType); ok {
			return string(valuer)
		}
		return nil
	}, BackendType(""))
	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(FileBackendMode); ok {
			return string(valuer)
		}
		return nil
	}, FileBackendMode(""))

	err := v.Struct(c)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		logger.Get().Error("Config validation failed",
			"exception", validationErrors,
		)
		return errors.New("config validation failed")
	}

	return nil
}

func Load(path string) (*Config, error) {
	config := &Config{}

	err := loadFromFile(path, config)
	if err != nil {
		logger.Get().Error("Failed to get config from file",
			"exception", err.Error(),
		)
		return nil, err
	}

	err = envconfig.Process("", config)
	if err != nil {
		logger.Get().Error("Failed to get config from environment variables",
			"exception", err.Error(),
		)
		return nil, err
	}

	config.setDefaults()

	err = config.validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadFromFile(path string, config *Config) error {
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			logger.Get().Error("Failed to open config file",
				"exception", err.Error(),
			)
			return err
		}
		defer file.Close()

		err = loadFromReader(config, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadFromReader(config *Config, reader io.Reader) error {
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(config)
	if err != nil {
		logger.Get().Error("Failed to decode config",
			"exception", err.Error(),
		)
		return err
	}
	return nil
}
