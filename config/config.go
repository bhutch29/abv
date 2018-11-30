package config

import (
	"github.com/spf13/viper"
	"github.com/mitchellh/go-homedir"
)

// New returns a copy of a preconfigured Viper configuration helper
func New() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	v.AddConfigPath(home + "/.abv/")
	v.AddConfigPath(".")

	v.WatchConfig()

	v.SetEnvPrefix("abv")
	v.BindEnv("untappdId")
	v.BindEnv("untappdSecret")

	// v.SetDefault("", "")

	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}
