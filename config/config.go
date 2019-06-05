// Package config handles the registration of environment variables and preferred filesystem paths.
package config

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var v *viper.Viper

// New returns a preconfigured Viper configuration helper
func New() (*viper.Viper, error) {
	if v == nil {
		return newViper()
	}
	return v, nil
}

// newViper collects config settings and returns a Viper object
func newViper() (*viper.Viper, error) {
	v = viper.New()
	v.SetConfigName("config")
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	v.AddConfigPath(path.Join(home, ".abv"))
	v.AddConfigPath(".")
	v.WatchConfig()

	v.SetEnvPrefix("abv")
	v.BindEnv("untappdId")
	v.BindEnv("untappdSecret")

	v.SetDefault("configPath", path.Join(home, ".abv"))
	v.SetDefault("webRoot", path.Join("/srv", "http"))
	v.SetDefault("apiUrl", "localhost")

	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
