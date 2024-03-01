package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/yangxikun/gsproxy"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("GSPROXY")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Fatal(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	flag.String("listen", ":8080", "proxy listen addr")
	flag.String("expose_metrics_listen", "", "expose metrics listen addr")
	flag.String("credentials", "", "basic credentials: username1:password1,username2:password2")
	flag.Bool("gen_credential", false, "generate a credential for auth")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	if err = viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatal(err)
	}
	var config struct {
		Listen              string
		ExposeMetricsListen string `mapstructure:"expose_metrics_listen"`
		Credentials         []string
		GenCredential       bool `mapstructure:"gen_credential"`
	}
	if err = viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}
	server := gsproxy.NewServer(config.Listen, config.ExposeMetricsListen, config.Credentials, config.GenCredential)
	server.Start()
}
