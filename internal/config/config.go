package config

import (
	"time"

	"github.com/tierklinik-dobersberg/apis/pkg/service"
)

type Config struct {
	service.MongoConfig
	service.BaseConfig

	DefaultInitialTimeRequirement    time.Duration `env:"INITIAL_TIME_REQUIREMENT,default=15m"`
	DefaultAdditionalTimeRequirement time.Duration `env:"ADDITIONAL_TIME_REQUIREMENT,default=10m"`
}
