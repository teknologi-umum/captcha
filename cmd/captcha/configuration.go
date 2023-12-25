package main

import (
	"errors"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type FeatureFlag struct {
	Analytics         bool `yaml:"analytics" json:"analytics" env:"FEATURE_FLAG_ANALYTICS"`
	BadwordsInsertion bool `yaml:"badwords_insertion" json:"badwords_insertion" env:"FEATURE_FLAG_BADWORDS_INSERTION"`
	Dukun             bool `yaml:"dukun" json:"dukun" env:"FEATURE_FLAG_DUKUN"`
	UnderAttack       bool `yaml:"under_attack" json:"under_attack" env:"FEATURE_FLAG_UNDER_ATTACK" env-default:"true"`
}

type Configuration struct {
	Environment string      `yaml:"environment" json:"environment" toml:"environment" env:"ENVIRONMENT" env-default:"production"`
	BotToken    string      `yaml:"bot_token" json:"bot_token" toml:"bot_token" env:"BOT_TOKEN" env-required:"true"`
	FeatureFlag FeatureFlag `yaml:"feature_flag" json:"feature_flag"`
	HomeGroupID int64       `yaml:"home_group_id" json:"home_group_id" env:"HOME_GROUP_ID"`
	AdminIds    []string    `yaml:"admin_ids" json:"admin_ids" env:"ADMIN_IDS"`
	SentryDSN   string      `yaml:"sentry_dsn" json:"sentry_dsn" env:"SENTRY_DSN"`
	Database    struct {
		PostgresUrl string `yaml:"postgres_url" json:"postgres_url" env:"POSTGRES_URL"`
		MongoUrl    string `yaml:"mongo_url" json:"mongo_url" env:"MONGO_URL"`
	} `yaml:"database" json:"database"`
	HTTPServer struct {
		ListeningHost string `yaml:"listening_host" json:"listening_host" env:"HTTP_HOST"`
		ListeningPort string `yaml:"listening_port" json:"listening_port" env:"HTTP_PORT" env-default:"8080"`
	}
	UnderAttack struct {
		DatastoreProvider string `yaml:"datastore_provider" json:"datastore_provider" env:"UNDER_ATTACK__DATASTORE_PROVIDER" env-default:"memory"`
	}
}

func ParseConfiguration(configurationFilePath string) (Configuration, error) {
	configuration := Configuration{}
	err := cleanenv.ReadConfig(configurationFilePath, &configuration)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Configuration{}, err
		}

		err := cleanenv.ReadEnv(&configuration)
		if err != nil {
			return Configuration{}, err
		}
	}

	return configuration, nil
}
