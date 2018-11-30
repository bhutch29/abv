package config

import (
	"github.com/spf13/viper"
	"github.com/mitchellh/go-homedir"
)

var v *viper.Viper

// New returns a preconfigured Viper configuration helper
func New() (*viper.Viper, error) {
	if v == nil {
		v = viper.New()
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

		v.SetDefault("imageCachePath", home + "/.abv/images")
		v.SetDefault("configPath", home + "/.abv")
		v.SetDefault("webRoot", "/srv/http")

		if err = v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	return v, nil
}